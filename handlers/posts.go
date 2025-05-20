package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	uploadDir     = "./static/uploads"
	maxUploadSize = 20 * 1024 * 1024 // 20MB
)

func init() {
	// Create upload directory if it doesn't exist
	os.MkdirAll(uploadDir, 0755)
}

type Post struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Content       string    `json:"content"`
	ImagePath     *string   `json:"image_path,omitempty"`
	Categories    []string  `json:"categories"`
	LikesCount    int       `json:"likes_count"`
	DislikesCount int       `json:"dislike_count"`
	CreatedAt     time.Time `json:"created_at"`
}

func CreatePostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check authentication first
		userID, err := getAuthenticatedUserID(db, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Variables to hold post data
		var content string
		var imagePath sql.NullString

		// Check content type
		contentType := r.Header.Get("Content-Type")

		if strings.HasPrefix(contentType, "application/json") {
			// JSON request (text-only post)
			var post struct {
				Content string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			content = post.Content
		} else if strings.HasPrefix(contentType, "multipart/form-data") {
			// Multipart form (possible file upload)
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				http.Error(w, "File too large or invalid form", http.StatusBadRequest)
				return
			}

			content = r.FormValue("content")
			file, fileHeader, err := r.FormFile("image")
			if err == nil {
				defer file.Close()

				// Validate file
				if fileHeader.Size > maxUploadSize {
					http.Error(w, "File too large (max 20MB)", http.StatusBadRequest)
					return
				}

				buff := make([]byte, 512)
				if _, err := file.Read(buff); err != nil {
					http.Error(w, "Invalid file", http.StatusBadRequest)
					return
				}
				if _, err := file.Seek(0, 0); err != nil {
					http.Error(w, "File error", http.StatusInternalServerError)
					return
				}

				filetype := http.DetectContentType(buff)
				if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/gif" {
					http.Error(w, "Only JPEG, PNG and GIF images are allowed", http.StatusBadRequest)
					return
				}

				// Generate unique filename
				ext := filepath.Ext(fileHeader.Filename)
				newFileName := uuid.New().String() + ext
				newFilePath := filepath.Join(uploadDir, newFileName)

				// Save file
				dst, err := os.Create(newFilePath)
				if err != nil {
					http.Error(w, "Failed to save file", http.StatusInternalServerError)
					return
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					http.Error(w, "Failed to save file", http.StatusInternalServerError)
					return
				}

				imagePath = sql.NullString{String: newFileName, Valid: true}
			} else if err != http.ErrMissingFile {
				http.Error(w, "Invalid file upload", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Unsupported content type", http.StatusBadRequest)
			return
		}

		// Validate at least content or image exists
		if content == "" && !imagePath.Valid {
			http.Error(w, "Post must have content or an image", http.StatusBadRequest)
			return
		}

		// Create post
		postID := uuid.New().String()
		_, err = db.Exec("INSERT INTO posts (id, user_id, content, image_path) VALUES (?, ?, ?, ?)",
			postID, userID, content, imagePath)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		// Handle categories if they exist
		if categories := r.Form["categories"]; len(categories) > 0 {
			for _, catID := range categories {
				_, err := db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
					postID, catID)
				if err != nil {
					log.Printf("Failed to add category %s to post %s: %v", catID, postID, err)

				}
			}
		}

		// Fetch the complete post data to return to client
		var createdPost Post
		var dbImagePath sql.NullString

		err = db.QueryRow(`
            SELECT p.id, p.user_id, u.username, p.content, p.image_path, p.created_at
            FROM posts p
            JOIN users u ON p.user_id = u.id
            WHERE p.id = ?`, postID).
			Scan(&createdPost.ID, &createdPost.UserID, &createdPost.Username,
				&createdPost.Content, &dbImagePath, &createdPost.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to fetch created post", http.StatusInternalServerError)
			return
		}

		// Convert NullString to *string
		if dbImagePath.Valid {
			createdPost.ImagePath = &dbImagePath.String
		}

		categories, categErrs := GetPostCategories(db, createdPost.ID)
		if categErrs != nil {
			log.Println("error getting categories from db", err)
			http.Error(w, "An error occured, kindly check back later", http.StatusInternalServerError)
		}

		createdPost.Categories = categories

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdPost)
	}
}

func GetPostCategories(db *sql.DB, id string) ([]string, error) {
	rows, err := db.Query(`
			SELECT c.name 
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
			`, id)
	if err != nil {

		return nil, err
	}
	defer rows.Close()

	categories := []string{}
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			log.Fatalf("Error scanning category: %v", err)
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func GetReactionCountsForPost(db *sql.DB, id string) (int, int, error) {
	query := `
		SELECT COUNT(*) FILTER (WHERE r.type = 'like'), COUNT(*) FILTER (WHERE r.type = 'dislike')
		FROM reactions r
		WHERE r.post_id = ?
	`
	var likeCount, dislikeCount int
	err := db.QueryRow(query, id).Scan(&likeCount, &dislikeCount)
	if err != nil {
		return 0, 0, err
	}
	return likeCount, dislikeCount, nil
}

func ListPostsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
			SELECT p.id, p.user_id, u.username, p.content, p.image_path, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			ORDER BY p.created_at DESC
			LIMIT 50`)
		if err != nil {
			log.Println("list post error", err)
			http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		posts := []Post{}
		for rows.Next() {
			var post Post
			var imagePath sql.NullString // Use sql.NullString to handle NULL values

			err := rows.Scan(&post.ID, &post.UserID, &post.Username,
				&post.Content, &imagePath, &post.CreatedAt)
			if err != nil {
				log.Println("add post to array error", err)
				http.Error(w, "Failed to read posts", http.StatusInternalServerError)
				return
			}

			// Convert NullString to *string
			// in templates the *string will be implicitly derefrenced
			if imagePath.Valid {
				post.ImagePath = &imagePath.String
			} else {
				post.ImagePath = nil
			}

			categories, categErrs := GetPostCategories(db, post.ID)
			if categErrs != nil {
				log.Println("error getting categories from db", err)
				http.Error(w, "An error occured, kindly check back later", http.StatusInternalServerError)
			}

			post.Categories = categories

			posts = append(posts, post)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(posts)
	}
}

// Helper function to get authenticated user ID
func getAuthenticatedUserID(db *sql.DB, r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", err
	}

	var userID string
	err = db.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).
		Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

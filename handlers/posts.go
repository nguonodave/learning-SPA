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
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	ImagePath *string   `json:"image_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
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

		// Fetch the complete post data to return to client
		var createdPost struct {
			ID        string    `json:"id"`
			UserID    string    `json:"user_id"`
			Username  string    `json:"username"`
			Content   string    `json:"content"`
			ImagePath *string   `json:"image_path,omitempty"`
			CreatedAt time.Time `json:"created_at"`
		}
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdPost)
	}
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

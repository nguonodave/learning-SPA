package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	ImagePath string    `json:"image_path,omitempty"`
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

		// Check if this is a multipart form (file upload)
		var content string
		var imagePath string

		if r.Header.Get("Content-Type") == "application/json" {
			// JSON request (text-only post)
			var post struct {
				Content string `json:"content"`
			}
			err = json.NewDecoder(r.Body).Decode(&post)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			content = post.Content
		} else {
			// Multipart form (possible file upload)
			err := r.ParseMultipartForm(maxUploadSize)
			if err != nil {
				http.Error(w, "File too large or invalid form", http.StatusBadRequest)
				return
			}

			content = r.FormValue("content")
			file, fileHeader, err := r.FormFile("image")
			if err == nil {
				defer file.Close()

				// Validate file size
				if fileHeader.Size > maxUploadSize {
					http.Error(w, "File too large (max 20MB)", http.StatusBadRequest)
					return
				}

				// Validate file type
				buff := make([]byte, 512)
				_, err = file.Read(buff)
				if err != nil {
					http.Error(w, "Invalid file", http.StatusBadRequest)
					return
				}
				file.Seek(0, 0) // Reset file pointer

				filetype := http.DetectContentType(buff)
				if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/gif" {
					http.Error(w, "Only JPEG, PNG and GIF images are allowed", http.StatusBadRequest)
					return
				}

				// Create a new file name
				ext := filepath.Ext(fileHeader.Filename)
				newFileName := uuid.New().String() + ext
				newFilePath := filepath.Join(uploadDir, newFileName)

				// Create the file
				dst, err := os.Create(newFilePath)
				if err != nil {
					http.Error(w, "Failed to save file", http.StatusInternalServerError)
					return
				}
				defer dst.Close()

				// Copy the uploaded file to the destination
				_, err = io.Copy(dst, file)
				if err != nil {
					http.Error(w, "Failed to save file", http.StatusInternalServerError)
					return
				}

				imagePath = newFileName
			} else if err != http.ErrMissingFile {
				http.Error(w, "Invalid file upload", http.StatusBadRequest)
				return
			}
		}

		// Basic validation
		if len(content) == 0 && imagePath == "" {
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

		// Return the created post
		var createdPost Post
		err = db.QueryRow(`
			SELECT p.id, p.user_id, u.username, p.content, p.image_path, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.id = ?`, postID).
			Scan(&createdPost.ID, &createdPost.UserID, &createdPost.Username,
				&createdPost.Content, &createdPost.ImagePath, &createdPost.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to fetch created post", http.StatusInternalServerError)
			return
		}

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
			err := rows.Scan(&post.ID, &post.UserID, &post.Username,
				&post.Content, &post.ImagePath, &post.CreatedAt)
			if err != nil {
				log.Println("add post to array error", err)
				http.Error(w, "Failed to read posts", http.StatusInternalServerError)
				return
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

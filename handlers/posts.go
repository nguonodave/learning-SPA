package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

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

		var post struct {
			Content string `json:"content"`
		}
		err = json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Basic validation
		if len(post.Content) == 0 {
			http.Error(w, "Content cannot be empty", http.StatusBadRequest)
			return
		}

		// Create post
		postID := uuid.New().String()
		_, err = db.Exec("INSERT INTO posts (id, user_id, content) VALUES (?, ?, ?)",
			postID, userID, post.Content)
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

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func CreateCommentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getAuthenticatedUserID(db, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		postID := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		postID = strings.TrimSuffix(postID, "/comments")

		var request struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		request.Content = strings.TrimSpace(request.Content)
		if len(request.Content) == 0 {
			http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
			return
		}

		commentID := uuid.New().String()
		_, err = db.Exec(`
			INSERT INTO comments (id, post_id, user_id, content)
			VALUES (?, ?, ?, ?)`,
			commentID, postID, userID, request.Content)
		if err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}

		// Return the created comment
		var comment Comment
		err = db.QueryRow(`
			SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.id = ?`, commentID).
			Scan(&comment.ID, &comment.PostID, &comment.UserID,
				&comment.Username, &comment.Content, &comment.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to fetch comment", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
	}
}

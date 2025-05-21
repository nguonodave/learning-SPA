package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
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
		commentCount, commentCountErr := GetPostCommentCount(db, postID)
		if commentCountErr != nil {
			log.Fatalf("error in GetPostCommentCount: %v", commentCountErr)
			http.Error(w, "An error occured try again later", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(commentCount)
	}
}

func GetCommentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		postID := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		postID = strings.TrimSuffix(postID, "/comments")

		rows, err := db.Query(`
			SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.post_id = ?
			ORDER BY c.created_at DESC`, postID)
		if err != nil {
			http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var comments []Comment
		for rows.Next() {
			var comment Comment
			err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID,
				&comment.Username, &comment.Content, &comment.CreatedAt)
			if err != nil {
				http.Error(w, "Failed to read comments", http.StatusInternalServerError)
				return
			}
			comments = append(comments, comment)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}

func GetPostCommentCount(db *sql.DB, postID string) (int, error) {
	commentCount := 0
	err := db.QueryRow(`
			SELECT COUNT(*) as comment_count
			FROM comments
			WHERE post_id = ?`, postID).
		Scan(&commentCount)
	if err != nil {
		return 0, err
	}
	return commentCount, nil
}

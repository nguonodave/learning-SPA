package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func ReactToPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getAuthenticatedUserID(db, r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		postID := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		postID = strings.TrimSuffix(postID, "/react")

		var request struct {
			Type string `json:"type"` // "like" or "dislike"
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if request.Type != "like" && request.Type != "dislike" {
			http.Error(w, "Invalid reaction type", http.StatusBadRequest)
			return
		}

		// Check if reaction exists
		var existingType string
		err = db.QueryRow(`
			SELECT type FROM reactions 
			WHERE user_id = ? AND post_id = ? AND comment_id IS NULL`,
			userID, postID).Scan(&existingType)

		if err == nil {
			// Reaction exists - toggle if same type, update if different
			if existingType == request.Type {
				// Remove reaction
				_, err = db.Exec(`
					DELETE FROM reactions 
					WHERE user_id = ? AND post_id = ? AND comment_id IS NULL`,
					userID, postID)
			} else {
				// Update reaction
				_, err = db.Exec(`
					UPDATE reactions SET type = ?
					WHERE user_id = ? AND post_id = ? AND comment_id IS NULL`,
					request.Type, userID, postID)
			}
		} else if err == sql.ErrNoRows {
			// New reaction
			_, err = db.Exec(`
				INSERT INTO reactions (id, user_id, post_id, type)
				VALUES (?, ?, ?, ?)`,
				uuid.New().String(), userID, postID, request.Type)
		}

		if err != nil {
			http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
			return
		}

		// Return updated counts
		var counts struct {
			Likes    int `json:"likes"`
			Dislikes int `json:"dislikes"`
			UserVote int `json:"userVote"` // 1 for like, -1 for dislike, 0 for none
		}

		db.QueryRow(`
			SELECT COUNT(*) FROM reactions 
			WHERE post_id = ? AND type = 'like' AND comment_id IS NULL`, postID).
			Scan(&counts.Likes)

		db.QueryRow(`
			SELECT COUNT(*) FROM reactions 
			WHERE post_id = ? AND type = 'dislike' AND comment_id IS NULL`, postID).
			Scan(&counts.Dislikes)

		db.QueryRow(`
			SELECT CASE WHEN type = 'like' THEN 1 WHEN type = 'dislike' THEN -1 ELSE 0 END
			FROM reactions 
			WHERE user_id = ? AND post_id = ? AND comment_id IS NULL`,
			userID, postID).Scan(&counts.UserVote)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(counts)
	}
}

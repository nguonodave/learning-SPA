package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ListCategoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query("SELECT id, name FROM categories ORDER BY name")
		if err != nil {
			http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []Category
		for rows.Next() {
			var cat Category
			if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
				http.Error(w, "Failed to read categories", http.StatusInternalServerError)
				return
			}
			categories = append(categories, cat)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	}
}

func GetCategoryPostsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		categoryID := strings.TrimPrefix(r.URL.Path, "/api/categories/")
		categoryID = strings.TrimSuffix(categoryID, "/posts")

		rows, err := db.Query(`
			SELECT p.id, p.user_id, u.username, p.content, p.image_path, p.created_at
			FROM posts p
			JOIN users u ON p.user_id = u.id
			JOIN post_categories pc ON p.id = pc.post_id
			WHERE pc.category_id = ?
			ORDER BY p.created_at DESC
			LIMIT 50`, categoryID)
		if err != nil {
			http.Error(w, "Failed to fetch category posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			var imagePath sql.NullString

			err := rows.Scan(&post.ID, &post.UserID, &post.Username,
				&post.Content, &imagePath, &post.CreatedAt)
			if err != nil {
				http.Error(w, "Failed to read posts", http.StatusInternalServerError)
				return
			}

			if imagePath.Valid {
				post.ImagePath = &imagePath.String
			}

			posts = append(posts, post)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}
}

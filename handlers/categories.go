package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
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

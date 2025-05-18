package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func SeedCategories(db *sql.DB) error {
	categories := []string{
		"Technology",
		"Travel",
		"Food",
		"Lifestyle",
		"Sports",
		"Music",
		"Art",
		"Science",
	}

	for _, category := range categories {
		_, err := db.Exec("INSERT OR IGNORE INTO categories (id, name) VALUES (?, ?)",
			uuid.New().String(), category)
		if err != nil {
			return fmt.Errorf("failed to seed category %s: %w", category, err)
		}
	}
	return nil
}

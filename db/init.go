// db/init.go
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens a connection to the DB and runs schema from schema/schema.sql
func InitDB(path string, schemaFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	// Load the schema from the .sql file
	schemaBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read schema file: %w", err)
	}

	// Execute schema
	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot execute schema: %w", err)
	}

	return db, nil
}

package main

import (
	"log"
	"net/http"
	"postSPA/db"
)

func main() {
    // Open or create database, and initialize schema
	sqlitePath := "app.db"
	schemaFile := "schema/schema.sql"

	database, initDbErr := db.InitDB(sqlitePath, schemaFile)
	if initDbErr != nil {
		log.Fatalf("DB init failed: %v", initDbErr)
	}
	defer database.Close()
	log.Println("✅ Database initialized and schema applied.")

    // Serve frontend (JS modules, HTML)
    http.Handle("/", http.FileServer(http.Dir("./frontend")))

    // Serve uploaded images
    http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./static/uploads"))))

    log.Println("Server started on http://localhost:8080")
    serveErr := http.ListenAndServe(":8080", nil)
    if serveErr != nil {
        log.Fatal(serveErr)
    }
}

package main

import (
	"log"
	"net/http"
	"postSPA/auth"
	"postSPA/db"
)

func main() {
	sqlitePath := "app.db"
	schemaFile := "schema/schema.sql"

	database, initDbErr := db.InitDB(sqlitePath, schemaFile)
	if initDbErr != nil {
		log.Fatalf("DB init failed: %v", initDbErr)
	}
	defer database.Close()
	log.Println("✅ Database initialized and schema applied.")

	// API handlers
	http.Handle("/api/register", auth.RegisterHandler(database))
	http.Handle("/api/login", auth.LoginHandler(database))

	// Frontend static files
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./static/uploads"))))

	log.Println("Server started on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

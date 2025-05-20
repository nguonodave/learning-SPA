package main

import (
	"log"
	"net/http"
	"postSPA/db"
	"postSPA/handlers"
	"strings"
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

	if err := db.SeedCategories(database); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}
	log.Println("✅ Categories seeded")

	// Auth handlers
	http.HandleFunc("/api/register", handlers.RegisterHandler(database))
	http.HandleFunc("/api/login", handlers.LoginHandler(database))
	http.HandleFunc("/api/logout", handlers.LogoutHandler(database))
	http.HandleFunc("/api/check-auth", handlers.AuthCheckHandler(database))
	http.HandleFunc("/api/posts", handlers.ListPostsHandler(database))
	http.HandleFunc("/api/posts/create", handlers.CreatePostHandler(database))
	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/posts") {
			handlers.GetCategoryPostsHandler(database)(w, r)
		} else {
			handlers.ListCategoriesHandler(database)(w, r)
		}
	})
	http.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/react") {
			handlers.ReactToPostHandler(database)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/categories") {
			handlers.GetPostCategoriesHandler(database)(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

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

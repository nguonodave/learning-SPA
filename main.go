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

	initDbErr := db.InitDB(sqlitePath, schemaFile)
	if initDbErr != nil {
		log.Fatalf("DB init failed: %v", initDbErr)
	}
	defer db.Db.Close()
	log.Println("✅ Database initialized and schema applied.")

	if err := db.SeedCategories(db.Db); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}
	log.Println("✅ Categories seeded")

	// Auth handlers
	http.HandleFunc("/api/register", handlers.RegisterHandler(db.Db))
	http.HandleFunc("/api/login", handlers.LoginHandler(db.Db))
	http.HandleFunc("/api/logout", handlers.LogoutHandler(db.Db))
	http.HandleFunc("/api/check-auth", handlers.AuthCheckHandler(db.Db))
	http.HandleFunc("/api/posts", handlers.ListPostsHandler(db.Db))
	http.HandleFunc("/api/posts/create", handlers.CreatePostHandler(db.Db))
	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/api/categories" || path == "/api/categories/":
			handlers.ListCategoriesHandler(db.Db)(w, r)
		case strings.HasSuffix(path, "/posts"):
			handlers.GetCategoryPostsHandler(db.Db)(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	http.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/react"):
			handlers.ReactToPostHandler(db.Db)(w, r)
		case strings.HasSuffix(r.URL.Path, "/comments"):
			if r.Method == http.MethodPost {
				handlers.CreateCommentHandler(db.Db)(w, r)
			} else {
				handlers.GetCommentsHandler(db.Db)(w, r)
			}
		default:
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

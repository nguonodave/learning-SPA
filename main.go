package main

import (
    "log"
    "net/http"
)

func main() {
    // Serve frontend (JS modules, HTML)
    http.Handle("/", http.FileServer(http.Dir("./frontend")))

    // Serve uploaded images
    http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./static/uploads"))))

    log.Println("Server started on http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
}

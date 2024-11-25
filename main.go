package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Initialize MongoDB
	if err := initializeDatabase(); err != nil {
		fmt.Println("Failed to connect to MongoDB:", err)
		return
	}

	// Start background tasks
	startBackgroundTasks()

	// Serve static files and handle routes
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/shorten", shortenURLHandler)

	// Register routes
	http.HandleFunc("/delete-url", deleteURLHandler) // Use the handler from handler.go
	http.HandleFunc("/check-url-or-alias", checkURLOrAliasHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

package main

import (
	"log"
	"net/http"
	"backend/api" 
)

func main() {
	// API routes
	http.HandleFunc("/api/courses", api.Handler)

	port := ":8000"
	log.Printf("🚀 Local server running on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

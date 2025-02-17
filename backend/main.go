package main

import (
	"log"
	"net/http"
	"backend/api" // ✅ Import the API package
)

func main() {
	// Define API routes
	http.HandleFunc("/api/courses", api.Handler)

	port := ":3000"
	log.Printf("🚀 Local server running on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

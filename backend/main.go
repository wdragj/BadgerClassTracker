package main

import (
	"backend/api/courses"
	checkAvailability "backend/api/cron/check-availability"
	"backend/api/register"
	"backend/api/subscribe"
	"backend/api/subscriptions"
	"backend/api/unsubscribe"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func init() {
	// Loads variables from a .env file into Go's environment.
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Println("No .env.local file found or error loading it.")
	}
  }

func main() {
	// API routes
	http.HandleFunc("/api/courses", courses.Handler)
	http.HandleFunc("/api/register", register.Handler)
	http.HandleFunc("/api/subscribe", subscribe.Handler)
	http.HandleFunc("/api/unsubscribe", unsubscribe.Handler)
	http.HandleFunc("/api/subscriptions", subscriptions.Handler)
	http.HandleFunc("/api/cron/check-availability", checkAvailability.Handler)


	port := ":8000"
	log.Printf("🚀 Local server running on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

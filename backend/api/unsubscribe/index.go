package unsubscribe

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UnsubscribePayload defines the JSON structure for unsubscription.
type UnsubscribePayload struct {
	UserEmail         string `json:"userEmail"`
	CourseID          string `json:"courseId"`
	CourseSubjectCode string `json:"courseSubjectCode"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload UnsubscribePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	log.Printf("Received unsubscribe payload: %+v\n", payload)

	dbURL := os.Getenv("POSTGRES_URL")
	pool, err := pgxpool.New(r.Context(), dbURL)
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		log.Println("DB connection error:", err)
		return
	}
	defer pool.Close()

	// Delete the subscription record for this user and course
	query := `
	DELETE FROM subscriptions
	WHERE user_id = (SELECT id FROM users WHERE email=$1)
	  AND course_id = $2
	  AND course_subject_code = $3
	`

	_, err = pool.Exec(r.Context(), query, payload.UserEmail, payload.CourseID, payload.CourseSubjectCode)
	if err != nil {
		http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
		log.Println("DB delete error:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Unsubscription successful"))
}

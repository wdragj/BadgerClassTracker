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

	// Delete the subscription record for this user and course.
	deleteQuery := `
		DELETE FROM subscriptions
		WHERE user_id = (SELECT id FROM users WHERE email=$1)
		  AND course_id = $2
		  AND course_subject_code = $3
	`
	_, err = pool.Exec(r.Context(), deleteQuery, payload.UserEmail, payload.CourseID, payload.CourseSubjectCode)
	if err != nil {
		http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
		log.Println("DB delete error:", err)
		return
	}

	// Now check if any subscriptions remain for this course.
	cleanupQuery := `
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE course_id = $1 AND course_subject_code = $2
	`
	var count int
	err = pool.QueryRow(r.Context(), cleanupQuery, payload.CourseID, payload.CourseSubjectCode).Scan(&count)
	if err != nil {
		log.Printf("Error checking remaining subscriptions for course %s: %v\n", payload.CourseID, err)
	} else if count == 0 {
		// No remaining subscriptions, so delete the course_availability record.
		deleteAvailabilityQuery := `
			DELETE FROM course_availability 
			WHERE course_id = $1 AND course_subject_code = $2
		`
		_, err = pool.Exec(r.Context(), deleteAvailabilityQuery, payload.CourseID, payload.CourseSubjectCode)
		if err != nil {
			log.Printf("Error deleting course_availability for course %s: %v\n", payload.CourseID, err)
		} else {
			log.Printf("Deleted course_availability record for course %s as no subscriptions remain.\n", payload.CourseID)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Unsubscription successful"))
}

package subscribe

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionPayload defines the JSON structure expected from the frontend.
type SubscriptionPayload struct {
	UserEmail         string `json:"userEmail"`
	UserFullName      string `json:"userFullName"`
	CourseID          string `json:"courseId"`
	CourseName        string `json:"courseName"`
	CourseSubjectCode string `json:"courseSubjectCode"`
	Credits           int    `json:"credits"`
	Title             string `json:"title"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers.
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

	var payload SubscriptionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	log.Printf("Received subscription payload: %+v\n", payload)

	// Connect to the database using the POSTGRES_URL environment variable.
	dbURL := os.Getenv("POSTGRES_URL")
	pool, err := pgxpool.New(r.Context(), dbURL)
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		log.Println("DB connection error:", err)
		return
	}
	defer pool.Close()

	now := time.Now()

	// Insert a new subscription record.
	// Since the last_checked column has been dropped, we only reference created_at.
	query := `
	INSERT INTO subscriptions (
	  user_id, user_email, user_fullname, course_id, 
	  course_name, course_subject_code, created_at,
	  credits, title
	)
	VALUES (
	  (SELECT id FROM users WHERE email=$1), 
	  $1, $2, $3, 
	  $4, $5, $6,
	  $7, $8
	)
	ON CONFLICT (user_id, course_id, course_subject_code)
	DO UPDATE SET
	  user_fullname = EXCLUDED.user_fullname,
	  course_name = EXCLUDED.course_name,
	  credits = EXCLUDED.credits,
	  title = EXCLUDED.title
	`
	_, err = pool.Exec(
		r.Context(),
		query,
		payload.UserEmail,         // $1
		payload.UserFullName,      // $2
		payload.CourseID,          // $3
		payload.CourseName,        // $4
		payload.CourseSubjectCode, // $5
		now,                       // $6 (for created_at)
		payload.Credits,           // $7
		payload.Title,             // $8
	)
	if err != nil {
		http.Error(w, "Failed to subscribe", http.StatusInternalServerError)
		log.Println("DB insert/upsert error:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Subscription successful"))
}

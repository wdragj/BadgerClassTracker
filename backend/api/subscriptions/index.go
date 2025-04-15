package subscriptions

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Subscription represents a subscription record returned to the frontend.
type Subscription struct {
	CourseID          string `json:"courseId"`
	CourseSubjectCode string `json:"courseSubjectCode"`
	CourseName        string `json:"courseName"`
	Credits           int    `json:"credits"`
	Title             string `json:"title"`
}

// SubscriptionsResponse wraps the subscriptions array.
type SubscriptionsResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get the userEmail from the query parameters
	userEmail := r.URL.Query().Get("userEmail")
	if userEmail == "" {
		http.Error(w, "userEmail query parameter is required", http.StatusBadRequest)
		return
	}

	// Connect to the database using DATABASE_URL env variable
	dbURL := os.Getenv("POSTGRES_URL")
	pool, err := pgxpool.New(r.Context(), dbURL)
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		log.Println("DB connection error:", err)
		return
	}
	defer pool.Close()

	// Query subscriptions for the user
	// Assuming your subscriptions table uses a composite unique key (user_id, course_id, course_subject_code)
	query := `
		SELECT course_id, course_subject_code, course_name, credits, title
		FROM subscriptions
		WHERE user_id = (SELECT id FROM users WHERE email = $1)
	`
	rows, err := pool.Query(r.Context(), query, userEmail)
	if err != nil {
		http.Error(w, "Failed to fetch subscriptions", http.StatusInternalServerError)
		log.Println("DB query error:", err)
		return
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(
			&sub.CourseID,
			&sub.CourseSubjectCode,
			&sub.CourseName,
			&sub.Credits,
			&sub.Title,
		); err != nil {
			http.Error(w, "Failed to scan subscription", http.StatusInternalServerError)
			log.Println("DB scan error:", err)
			return
		}
		subscriptions = append(subscriptions, sub)
	}

	response := SubscriptionsResponse{
		Subscriptions: subscriptions,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package register

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPayload struct {
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Image     string `json:"image,omitempty"`
	GoogleSub string `json:"googleSub,omitempty"`
}

type RequestBody struct {
	User UserPayload `json:"user"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	log.Printf("Received user data: %+v\n", req.User)

	dbURL := os.Getenv("POSTGRES_URL")
	pool, err := pgxpool.New(r.Context(), dbURL)
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		log.Println("DB connection error:", err)
		return
	}
	defer pool.Close()

	now := time.Now()

	query := `
	  INSERT INTO users (email, google_sub, name, image, created_at, last_logged_in)
	  VALUES ($1, $2, $3, $4, $5, $5)
	  ON CONFLICT (email)
	  DO UPDATE SET
	    google_sub      = EXCLUDED.google_sub,
	    name            = EXCLUDED.name,
	    image           = EXCLUDED.image,
	    last_logged_in  = EXCLUDED.last_logged_in
	  RETURNING id
	`

	// Use 'EXCLUDED.last_logged_in = now' so each new sign-in updates last_logged_in
	// for that user row. 'created_at' remains the original insertion time.
	// If you want to see the inserted/updated user ID, you can scan it here.

	_, err = pool.Exec(r.Context(), query,
		req.User.Email,
		req.User.GoogleSub,
		req.User.Name,
		req.User.Image,
		now, // used for both created_at (initial insert) and last_logged_in
	)

	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		log.Println("DB insert/upsert error:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User stored successfully."))
}

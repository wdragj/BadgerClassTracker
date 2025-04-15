package checkAvailability

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	// "time"

	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

// getRandomUserAgent returns a random user agent string.
func getRandomUserAgent() string {
	return uarand.GetRandom()
}

// Term holds term code and short description (unused below, but can be useful).
type Term struct {
	TermCode         string
	ShortDescription string
}

// checkClassStatus checks whether a course (by its name) is available.
func checkClassStatus(courseName string) (bool, error) {
	termCode := os.Getenv("TERM_CODE")
	if termCode == "" {
		termCode = "1262"
	}

	client := resty.New()

	payload := map[string]interface{}{
		"selectedTerm": termCode,
		"queryString":  courseName,
		"filters": []map[string]interface{}{
			{
				"has_child": map[string]interface{}{
					"type": "enrollmentPackage",
					"query": map[string]interface{}{
						"bool": map[string]interface{}{
							"must": []map[string]interface{}{
								{"match": map[string]interface{}{"packageEnrollmentStatus.status": "OPEN WAITLISTED"}},
								{"match": map[string]interface{}{"published": true}},
							},
						},
					},
				},
			},
		},
		"page":      1,
		"pageSize":  1,
		"sortOrder": "SCORE",
	}

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Accept":       "application/json, text/plain, */*",
			"Content-Type": "application/json",
			"User-Agent":   getRandomUserAgent(),
			"Origin":       "https://public.enroll.wisc.edu",
			"Referer":      fmt.Sprintf("https://public.enroll.wisc.edu/search?term=%s&keywords=%s", termCode, courseName),
		}).
		SetBody(payload).
		Post(apiURL)
	if err != nil {
		return false, err
	}

	if resp.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("API request failed with status: %d", resp.StatusCode())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return false, err
	}

	hits, ok := result["hits"].([]interface{})
	if !ok || len(hits) == 0 {
		return false, nil // No hits means the course is not available
	}

	// Check if any hit matches the course name (or your criteria)
	for _, h := range hits {
		hit, _ := h.(map[string]interface{})
		if designation, exists := hit["courseDesignation"].(string); exists && designation == courseName {
			return true, nil
		}
	}

	return false, nil
}

// sendGmailSMTP sends an email using Gmail’s SMTP servers.
// It uses net/smtp with an App Password (GMAIL_SMTP_PASS) instead of your real Google password.
func sendGmailSMTP(recipientEmail, term, courseName, prevStatus, newStatus string) error {
	// Get your Gmail address and app password from environment variables
	smtpEmail := os.Getenv("GMAIL_SMTP_EMAIL") // e.g., youremail@gmail.com
	smtpPass := os.Getenv("GMAIL_SMTP_PASS")   // 16-character app password

	if smtpEmail == "" || smtpPass == "" {
		return fmt.Errorf("GMAIL_SMTP_EMAIL or GMAIL_SMTP_PASS not set in environment")
	}

	// Gmail SMTP details
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	subject := fmt.Sprintf("Course Update: %s is now %s", courseName, newStatus)
	htmlBody := fmt.Sprintf("<p>%s<br><br>%s was previously <strong>%s</strong>.<br>It is now <strong>%s</strong>.<br><br>Thank you.</p>",
		term, courseName, prevStatus, newStatus)

	// Build the raw MIME message.
	msg := strings.Join([]string{
		"To: " + recipientEmail,
		"From: " + smtpEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		htmlBody,
	}, "\r\n")

	// Set up authentication using your app password.
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, smtpHost)

	// Actually send the email.
	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, []string{recipientEmail}, []byte(msg)); err != nil {
		return err
	}

	log.Printf("Email sent to %s via Gmail SMTP\n", recipientEmail)
	return nil
}

// Handler is the HTTP handler for the cron job.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for the cron job.
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	dbURL := os.Getenv("POSTGRES_URL")
	pool, err := pgxpool.New(r.Context(), dbURL)
	if err != nil {
		http.Error(w, "Failed to connect to DB", http.StatusInternalServerError)
		log.Println("DB connection error:", err)
		return
	}
	defer pool.Close()

	// Query distinct courses from subscriptions.
	query := `
		SELECT DISTINCT course_name, course_id, course_subject_code
		FROM subscriptions
	`
	rows, err := pool.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to query subscriptions", http.StatusInternalServerError)
		log.Println("Query error:", err)
		return
	}
	defer rows.Close()

	type CourseInfo struct {
		CourseName        string
		CourseID          string
		CourseSubjectCode string
	}
	var coursesToCheck []CourseInfo
	for rows.Next() {
		var info CourseInfo
		if err := rows.Scan(&info.CourseName, &info.CourseID, &info.CourseSubjectCode); err != nil {
			http.Error(w, "Failed to scan course info", http.StatusInternalServerError)
			log.Println("Scan error:", err)
			return
		}
		coursesToCheck = append(coursesToCheck, info)
	}

	termShortDesc := os.Getenv("TERM_SHORT_DESCRIPTION")
	if termShortDesc == "" {
		termShortDesc = "Term 1262"
	}

	// For each course, check availability and update the centralized course_availability table.
	for _, course := range coursesToCheck {
		available, err := checkClassStatus(course.CourseName)
		if err != nil {
			log.Printf("Error checking %s: %v\n", course.CourseName, err)
			continue
		}

		// Determine new status.
		newStatus := "full"
		if available {
			newStatus = "open"
		}

		// Retrieve previous status from course_availability.
		var prevStatus string
		statusQuery := `
			SELECT course_status
			FROM course_availability
			WHERE course_id = $1 AND course_subject_code = $2
		`
		err = pool.QueryRow(r.Context(), statusQuery, course.CourseID, course.CourseSubjectCode).Scan(&prevStatus)
		if err != nil {
			// If no record exists, assume default previous status as "full".
			prevStatus = "full"
		}

		// Upsert the centralized course availability record.
		upsertQuery := `
			INSERT INTO course_availability (course_id, course_subject_code, course_name, course_status, last_checked)
			VALUES ($1, $2, $3, $4, now())
			ON CONFLICT (course_id, course_subject_code)
			DO UPDATE SET course_status = EXCLUDED.course_status, last_checked = EXCLUDED.last_checked
		`
		_, err = pool.Exec(r.Context(), upsertQuery, course.CourseID, course.CourseSubjectCode, course.CourseName, newStatus)
		if err != nil {
			log.Printf("Error upserting course %s: %v\n", course.CourseName, err)
			continue
		}

		// If status changed, send notifications.
		if newStatus != prevStatus {
			emailQuery := `
				SELECT user_email
				FROM subscriptions
				WHERE course_id = $1 AND course_subject_code = $2
			`
			emailRows, err := pool.Query(r.Context(), emailQuery, course.CourseID, course.CourseSubjectCode)
			if err != nil {
				log.Printf("Error fetching subscribers for %s: %v\n", course.CourseName, err)
				continue
			}
			for emailRows.Next() {
				var email string
				if err := emailRows.Scan(&email); err != nil {
					log.Println("Error scanning email:", err)
					continue
				}
				// Send the email using Gmail SMTP
				if err := sendGmailSMTP(email, termShortDesc, course.CourseName, prevStatus, newStatus); err != nil {
					log.Printf("Error sending email to %s: %v\n", email, err)
				} else {
					log.Printf("Notification sent to %s for course %s\n", email, course.CourseName)
				}
			}
			emailRows.Close()
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Course availability check completed"))
}

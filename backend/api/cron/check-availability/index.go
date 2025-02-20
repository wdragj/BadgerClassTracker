package checkAvailability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailersend/mailersend-go"
)

const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

// getRandomUserAgent returns a random user agent.
func getRandomUserAgent() string {
	return uarand.GetRandom()
}

// checkClassStatus checks whether a course (by its name) is available.
func checkClassStatus(courseName string) (bool, error) {
	client := resty.New()

	payload := map[string]interface{}{
		"selectedTerm": "1254",
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
			"Referer":      fmt.Sprintf("https://public.enroll.wisc.edu/search?term=1254&keywords=%s", courseName),
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

// sendMailerSendEmail uses the MailerSend Go SDK to send an email notification.
func sendMailerSendEmail(recipientEmail, courseName, prevStatus, newStatus string) error {
	apiKey := os.Getenv("MAILERSEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("MAILERSEND_API_KEY not set")
	}

	// Initialize the MailerSend client.
	ms := mailersend.NewMailersend(apiKey)

	// Create a context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subject := fmt.Sprintf("Course Update: %s is now %s", courseName, newStatus)
	text := fmt.Sprintf("%s was previously %s.\nIt is now %s.\n\nThank you.", courseName, prevStatus, newStatus)
	html := fmt.Sprintf("<p>%s was previously <strong>%s</strong>.<br>It is now <strong>%s</strong>.<br><br>Thank you.</p>", courseName, prevStatus, newStatus)

	// Set sender details from environment variables.
	from := mailersend.From{
		Name:  os.Getenv("EMAIL_FROM_NAME"), // e.g., "Your Name"
		Email: os.Getenv("EMAIL_FROM"),      // e.g., "info@yourdomain.com"
	}

	// Prepare the recipient. If you have a name, you can include it; otherwise, leave it empty.
	recipients := []mailersend.Recipient{
		{
			Email: recipientEmail,
		},
	}

	message := ms.Email.NewMessage()
	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetText(text)
	message.SetHTML(html)

	// Send the email.
	_, err := ms.Email.Send(ctx, message)
	return err
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

	// Query distinct course names, course IDs, and subject codes from subscriptions.
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

	// For each course, check availability and update the subscription record.
	for _, course := range coursesToCheck {
		available, err := checkClassStatus(course.CourseName)
		if err != nil {
			log.Printf("Error checking %s: %v\n", course.CourseName, err)
			continue
		}

		// Get the previous status for the course.
		var prevStatus string
		statusQuery := `
			SELECT course_status
			FROM subscriptions
			WHERE course_id = $1 AND course_subject_code = $2
			LIMIT 1
		`
		err = pool.QueryRow(r.Context(), statusQuery, course.CourseID, course.CourseSubjectCode).Scan(&prevStatus)
		if err != nil {
			log.Printf("Error retrieving previous status for %s: %v\n", course.CourseName, err)
			// Proceed with update if status cannot be determined.
		}

		// Determine new status based on availability.
		newStatus := "full"
		if available {
			newStatus = "open"
		}

		// Always update last_checked. If the status changed, update it too.
		if newStatus != prevStatus {
			// Update both last_checked and course_status.
			updateQuery := `
				UPDATE subscriptions
				SET last_checked = now(), course_status = $1
				WHERE course_id = $2 AND course_subject_code = $3
			`
			_, err = pool.Exec(r.Context(), updateQuery, newStatus, course.CourseID, course.CourseSubjectCode)
			if err != nil {
				log.Printf("Error updating course %s: %v\n", course.CourseName, err)
			} else {
				log.Printf("Updated course %s from status %s to %s\n", course.CourseName, prevStatus, newStatus)
			}
		} else {
			// Only update last_checked if the status remains the same.
			updateQuery := `
				UPDATE subscriptions
				SET last_checked = now()
				WHERE course_id = $1 AND course_subject_code = $2
			`
			_, err = pool.Exec(r.Context(), updateQuery, course.CourseID, course.CourseSubjectCode)
			if err != nil {
				log.Printf("Error updating last_checked for course %s: %v\n", course.CourseName, err)
			} else {
				log.Printf("Updated last_checked for course %s (status remains %s)\n", course.CourseName, newStatus)
			}
		}

		// If the status changed, then send notifications.
		if newStatus != prevStatus {
			// Query subscriber emails for this course.
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
				if err := sendMailerSendEmail(email, course.CourseName, prevStatus, newStatus); err != nil {
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

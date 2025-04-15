package checkAvailability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	// "strconv"
	// "strings"
	"time"

	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailersend/mailersend-go"
)

const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

func getRandomUserAgent() string {
	return uarand.GetRandom()
}

// Term holds term code and short description.
type Term struct {
	TermCode         string
	ShortDescription string
}

// checkClassStatus checks whether a course (by its name) is available.
func checkClassStatus(courseName string) (bool, error) {
	// 1. Read from environment variables:
	termCode := os.Getenv("TERM_CODE")
	if termCode == "" {
		termCode = "1262"
	}
	// We only need the code here for the API query; description is not used in the payload.
	// But if you want it for logging, you can read it as well:
	// termShortDesc := os.Getenv("TERM_SHORT_DESCRIPTION")
	// if termShortDesc == "" {
	//     termShortDesc = "Term 1262"
	// }

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

// sendMailerSendEmail uses the MailerSend Go SDK to send an email notification.
// The email will include the term info on the first line.
func sendMailerSendEmail(recipientEmail, term, courseName, prevStatus, newStatus string) error {
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
	text := fmt.Sprintf("%s\n\n%s was previously %s.\nIt is now %s.\n\nThank you.", term, courseName, prevStatus, newStatus)
	html := fmt.Sprintf("<p>%s<br><br>%s was previously <strong>%s</strong>.<br>It is now <strong>%s</strong>.<br><br>Thank you.</p>", term, courseName, prevStatus, newStatus)

	from := mailersend.From{
		Name:  os.Getenv("EMAIL_FROM_NAME"),
		Email: os.Getenv("EMAIL_FROM"),
	}

	recipients := []mailersend.Recipient{
		{Email: recipientEmail},
	}

	message := ms.Email.NewMessage()
	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetText(text)
	message.SetHTML(html)

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

	// 1. Get environment variable for short description
	// (Term code not needed here unless you want to display it in the email body.)
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
				if err := sendMailerSendEmail(email, termShortDesc, course.CourseName, prevStatus, newStatus); err != nil {
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

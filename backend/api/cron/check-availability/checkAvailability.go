package checkAvailability

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

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

func Handler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for the cron job
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
		newStatus := "full"
		if available {
			newStatus = "open"
		}
		// Update last_checked and course_status for all subscriptions for this course.
		updateQuery := `
			UPDATE subscriptions
			SET last_checked = now(), course_status = $1
			WHERE course_id = $2 AND course_subject_code = $3
		`
		_, err = pool.Exec(r.Context(), updateQuery, newStatus, course.CourseID, course.CourseSubjectCode)
		if err != nil {
			log.Printf("Error updating course %s: %v\n", course.CourseName, err)
		} else {
			log.Printf("Updated course %s to status: %s\n", course.CourseName, newStatus)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Course availability check completed"))
}

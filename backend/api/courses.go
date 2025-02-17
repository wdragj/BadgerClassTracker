package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
)

// API URL for UW Madison Course Enrollment
const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

// Function to generate a random user-agent
func getRandomUserAgent() string {
	return uarand.GetRandom()
}

// Vercel Serverless Function Handler
func Handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		query = "*" // Default to fetching all courses
	}

	client := resty.New()
	payload := map[string]interface{}{
		"selectedTerm": "1254",
		"queryString":  query,
		"filters": []map[string]interface{}{
			{
				"has_child": map[string]interface{}{
					"type": "enrollmentPackage",
					"query": map[string]interface{}{
						"bool": map[string]interface{}{
							"must": []map[string]interface{}{
								{"match": map[string]interface{}{"published": true}},
							},
						},
					},
				},
			},
		},
		"page":      1,
		"pageSize":  10,
		"sortOrder": "SCORE",
	}

	// Make API request
	resp, err := client.R().
		SetHeaders(map[string]string{
			"Accept":       "application/json, text/plain, */*",
			"Content-Type": "application/json",
			"User-Agent":   getRandomUserAgent(),
			"Origin":       "https://public.enroll.wisc.edu",
			"Referer":      fmt.Sprintf("https://public.enroll.wisc.edu/search?term=1254&keywords=%s", query),
		}).
		SetBody(payload).
		Post(apiURL)

	if err != nil || resp.StatusCode() != http.StatusOK {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	hits, _ := result["hits"].([]interface{})
	var courses []map[string]interface{}
	for _, h := range hits {
		hit, _ := h.(map[string]interface{})
		courses = append(courses, map[string]interface{}{
			"termCode":     hit["termCode"],
			"subject":      hit["subject"].(map[string]interface{})["shortDescription"],
			"title":        hit["title"],
			"credits":      hit["creditRange"],
			"status":       hit["packageEnrollmentStatus"].(map[string]interface{})["status"],
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
)

// API URL for UW Madison Enrollment
const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

// Get a random user-agent
func getRandomUserAgent() string {
	return uarand.GetRandom()
}

// Fetch classes from the API
func fetchCourses(query string) (map[string]interface{}, error) {
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

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode())
	}

	// var result map[string]interface{}
	// err = json.Unmarshal(resp.Body(), &result)
	// if err != nil {
	// 	return nil, err
	// }

	// hits, ok := result["hits"].([]interface{})
	// if !ok || len(hits) == 0 {
	// 	return nil, nil
	// }

	// // Convert response to JSON
	// courses := []map[string]interface{}{}
	// for _, h := range hits {
	// 	hit, _ := h.(map[string]interface{})
	// 	courses = append(courses, hit)
	// }

	// return courses, nil
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 📌 **Handler for /api/courses**
func Handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	// Default to "*" if no query is provided
	if query == "" {
		query = "*"
	}

	courses, err := fetchCourses(query)
	if err != nil {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		log.Println("Error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

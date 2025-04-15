package courses

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/corpix/uarand"
	"github.com/go-resty/resty/v2"
)

const apiURL = "https://public.enroll.wisc.edu/api/search/v1"

// Term holds term code and short description.
type Term struct {
	TermCode         string `json:"termCode"`
	ShortDescription string `json:"shortDescription"`
}

func getRandomUserAgent() string {
	return uarand.GetRandom()
}

// expandSeasonAbbreviation expands abbreviated season names in the short description.
func expandSeasonAbbreviation(shortDesc string) string {
	replacements := map[string]string{
		"Sprng": "Spring",
		"Summr": "Summer",
	}
	for abbr, full := range replacements {
		shortDesc = strings.ReplaceAll(shortDesc, abbr, full)
	}
	return shortDesc
}

// fetchCourses queries the courses API
func fetchCourses(query string, page int, pageSize int, termCode string) (map[string]interface{}, error) {
	client := resty.New()

	payload := map[string]interface{}{
		"selectedTerm": termCode,
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
		"page":      page,
		"pageSize":  pageSize,
		"sortOrder": "SCORE",
	}

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Accept":       "application/json, text/plain, */*",
			"Content-Type": "application/json",
			"User-Agent":   getRandomUserAgent(),
			"Origin":       "https://public.enroll.wisc.edu",
			"Referer":      fmt.Sprintf("https://public.enroll.wisc.edu/search?term=%s&keywords=%s", termCode, query),
		}).
		SetBody(payload).
		Post(apiURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode())
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Handler is the API endpoint handler for /api/courses.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get the query parameter; default to "*" if not provided.
	query := r.URL.Query().Get("query")
	if query == "" {
		query = "*"
	}

	// Get page and pageSize from query params; default to 1 and 50 respectively.
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 50
	}

	// 1. Read environment variables for term code & short description.
	//    If they are empty, fall back to defaults.
	termCode := os.Getenv("TERM_CODE")
	if termCode == "" {
		termCode = "1262"
	}
	termShortDesc := os.Getenv("TERM_SHORT_DESCRIPTION")
	if termShortDesc == "" {
		termShortDesc = "Term 1262"
	}

	// 2. Create a Term struct from these variables.
	term := &Term{
		TermCode:         termCode,
		ShortDescription: termShortDesc,
	}

	// 3. Fetch courses using the environment-based term code.
	courses, err := fetchCourses(query, page, pageSize, term.TermCode)
	if err != nil {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		log.Println("Error fetching courses:", err)
		return
	}

	// 4. Combine courses and term info into a single response.
	response := map[string]interface{}{
		"term": map[string]interface{}{
			"termCode":         term.TermCode,
			"shortDescription": term.ShortDescription,
		},
		"courses": courses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

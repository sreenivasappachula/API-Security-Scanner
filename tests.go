package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func isIDParam(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "id")
}

func testIDOR(fullURL string, id int, authToken string) string {

	//println("[*] Testing IDOR on:", fullURL)

	if id <= 0 {
		return ""
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	url1 := buildPathURL(u, id)
	url2 := buildPathURL(u, id+1)

	r1, err1 := makeRequest(url1, authToken)
	if err1 != nil {
		return ""
	}
	defer r1.Body.Close()

	r2, err2 := makeRequest(url2, authToken)
	if err2 != nil {
		return ""
	}
	defer r2.Body.Close()

	// println("Checking:", url1, "=> %T", r1.StatusCode)
	// fmt.Printf("Type of r1.StatusCode: %T\n", r1.StatusCode)

	// println("Checking:", url2, "=> %T", r2.StatusCode)
	// fmt.Printf("Type of r1.StatusCode: %T\n", r2.StatusCode)
	// fmt.Printf("Condition: %v\n", r1.StatusCode == 200 && r2.StatusCode == 200)
	if r1.StatusCode == 200 && r2.StatusCode == 200 {
		return "Possible IDOR via path parameter"
	}

	return ""
}

func buildPathURL(u *url.URL, id int) string {
	copyURL := *u
	copyURL.RawQuery = ""

	path := strings.Trim(copyURL.Path, "/")
	if path == "" {
		copyURL.Path = "/" + strconv.Itoa(id)
		return copyURL.String()
	}

	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if _, err := strconv.Atoi(seg); err == nil {
			segments[i] = strconv.Itoa(id)
			copyURL.Path = "/" + strings.Join(segments, "/")
			return copyURL.String()
		}
	}

	segments = append(segments, strconv.Itoa(id))
	copyURL.Path = "/" + strings.Join(segments, "/")
	return copyURL.String()
}

func makeRequest(requestURL string, authToken string) (*http.Response, error) {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	// Set Authorization header if token is provided
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{}
	return client.Do(req)
}

func testAuth(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "No Authentication"
	}
	return ""
}

func testRateLimit(url string) string {
	for i := 0; i < 20; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			return ""
		}
	}
	return "No Rate Limiting"
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func readLineFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return []string{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	return lines
}

func testLogin_is_JWT(url string, paths []string) string {
	emails := readLineFromFile("emails.txt")
	passwords := readLineFromFile("passwords.txt")

	for _, path := range paths {
		if strings.Contains(strings.ToLower(path), "login") {
			for _, email := range emails {
				for _, password := range passwords {
					data := map[string]interface{}{
						"email":    email,
						"password": password,
					}

					jsonData, err := json.Marshal(data)
					if err != nil {
						fmt.Println(err)
						return ""
					}

					req, err := http.NewRequest("POST", url+path, bytes.NewBuffer(jsonData))
					if err != nil {
						fmt.Println(err)
						return ""
					}
					fmt.Println("Request body", string(jsonData))

					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:151.0) Gecko/20100101 Firefox/151.0")
					req.Header.Set("Referer", url+"/login")
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))
					req.Header.Set("Origin", url)
					req.Header.Set("Connection", "keep-alive")
					req.Header.Set("Cookie", "chat_session_id=6012c3e0-8f24-45d7-9765-db79ee63ce88")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println(err)
						return ""
					}
					defer resp.Body.Close()

					body, _ := io.ReadAll(resp.Body)
					fmt.Println("Status:", resp.Status)
					if resp.StatusCode == 200 {
						fmt.Println(string(body))
						return string(body)
					}
				}
			}
		}
	}
	return ""
}

// if response is 200 ok then
//check for authroization token exists or not

func testJWT(url string, paths []string) {
	// if authorization toke exists then  test for different use cases
}

func testIDOR(fullURL string, id int, authToken string) string {
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

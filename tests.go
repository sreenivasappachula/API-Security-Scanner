package main

import (
	"io"
	"net/http"
)

func testIDOR(baseURL, endpoint string) string {
	url1 := baseURL + endpoint + "/1"
	url2 := baseURL + endpoint + "/2"

	r1, err1 := http.Get(url1)
	r2, err2 := http.Get(url2)

	if err1 != nil || err2 != nil {
		return ""
	}
	defer r1.Body.Close()
	defer r2.Body.Close()

	b1, _ := io.ReadAll(r1.Body)
	b2, _ := io.ReadAll(r2.Body)

	if r1.StatusCode == 200 && r2.StatusCode == 200 && string(b1) != string(b2) {
		return "Possible IDOR"
	}
	return ""
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
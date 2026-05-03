package main

import (
	"io"
	"net/http"
	"regexp"
)

func fetchJSFiles(baseURL string) []string {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`src="(.*?\.js)"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	var jsFiles []string
	for _, m := range matches {
		js := m[1]
		if js[0] == '/' {
			jsFiles = append(jsFiles, baseURL+js)
		} else {
			jsFiles = append(jsFiles, js)
		}
	}
	return jsFiles
}

func extractEndpoints(jsURL string) []string {
	resp, err := http.Get(jsURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`/api/[a-zA-Z0-9_/\-]+`)
	return re.FindAllString(string(body), -1)
}

func discoverEndpoints(baseURL string) []string {
	endpointSet := make(map[string]bool)

	jsFiles := fetchJSFiles(baseURL)
	for _, js := range jsFiles {
		eps := extractEndpoints(js)
		for _, ep := range eps {
			endpointSet[ep] = true
		}
	}
	
	common := []string{"/api/users", "/api/login", "/api/admin"}
	for _, ep := range common {
		endpointSet[ep] = true
	}

	var endpoints []string
	for ep := range endpointSet {
		endpoints = append(endpoints, ep)
	}
	return endpoints
}
package main

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
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
		if strings.HasPrefix(js, "/") {
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

func readEndpointsFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	var endpoints []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			endpoints = append(endpoints, line)
		}
	}
	return endpoints
}

func discoverEndpoints(baseURL string) []string {
	endpointSet := make(map[string]bool)

	// 🔹 1. Read from file
	fileEndpoints := readEndpointsFromFile("apiendpoints_read.txt")
	for _, ep := range fileEndpoints {
		endpointSet[ep] = true
	}

	// 🔹 2. JS discovery (optional but useful)
	jsFiles := fetchJSFiles(baseURL)
	for _, js := range jsFiles {
		eps := extractEndpoints(js)
		for _, ep := range eps {
			endpointSet[ep] = true
		}
	}

	// Convert map → slice
	var endpoints []string
	for ep := range endpointSet {
		trimmed := strings.TrimSpace(ep)
		endpoints = append(endpoints, trimmed)
	}

	for ep := range endpoints {
		println("[*] Discovered:", endpoints[ep])
	}

	return endpoints
}

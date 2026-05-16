package endpoint

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func FetchJSFiles(baseURL string) []string {
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

func ExtractEndpoints(jsURL string) []string {
	resp, err := http.Get(jsURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`/api/[a-zA-Z0-9_/\-]+`)
	return re.FindAllString(string(body), -1)
}

func ReadEndpointsFromFile(filename string) []string {
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

func DiscoverEndpoints(baseURL string) []string {
	endpointSet := make(map[string]bool)

	// 🔹 1. Read from file
	fileEndpoints := ReadEndpointsFromFile("apiendpoints_read.txt")
	for _, ep := range fileEndpoints {
		endpointSet[ep] = true
	}

	// 🔹 2. JS discovery (optional but useful)
	jsFiles := FetchJSFiles(baseURL)
	for _, js := range jsFiles {
		eps := ExtractEndpoints(js)
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

func ParseURLAndPaths(filename string) (string, []string) {
	file, err := os.Open(filename)
	if err != nil {
		return "", []string{}
	}

	defer file.Close()
	var url string
	var paths []string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())
		//println("line : ", line)
		// Condition 1

		if strings.Contains(line, "/") && !strings.Contains(line, "//") {

			if strings.Contains(line, "json") {
				continue
			}
			path := line[1 : len(line)-4]
			paths = append(paths, path)
		}

		if strings.Contains(line, `"url": "`) {

			start := strings.Index(line, `" "url": "`) + len(`" "url": "`)
			end := strings.Index(line[start:], `"`)

			url = line[start-1 : start+end]

		}
	}
	if url == "http://localhost" || url == "https://localhost" {

		return url + ":8888", paths
	}
	return url, paths
}

package endpoint

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"regexp"
	"sort"
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

func ParseURLAndPaths(filename string) (string, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	type server struct {
		URL string `json:"url"`
	}

	type apiSpec struct {
		Servers []server               `json:"servers"`
		Paths   map[string]interface{} `json:"paths"`
	}

	var spec apiSpec
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&spec); err != nil {
		return "", nil, err
	}

	url := ""
	if len(spec.Servers) > 0 {
		url = strings.TrimSpace(spec.Servers[0].URL)
		parsed, err := neturl.Parse(url)
		if err == nil && parsed.Hostname() == "localhost" && parsed.Port() == "" {
			parsed.Host = net.JoinHostPort(parsed.Hostname(), "8888")
			url = parsed.String()
		}
	}

	paths := make([]string, 0, len(spec.Paths))
	for p := range spec.Paths {
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)

	return url, paths, nil
}

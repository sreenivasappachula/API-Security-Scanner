package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var inputURL string
var idorParam string
var authHeader string

func init() {
	flag.StringVar(&inputURL, "url", "", "Target URL (optional)")
	flag.StringVar(&idorParam, "idor_param", "", "ID parameter (e.g. id)")
	flag.StringVar(&authHeader, "auth_token", "", "Authorization token (e.g. your_token)")
}

func extractID(fullURL, param string) (int, string) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return 0, ""
	}

	val := u.Query().Get(param)
	if val == "" {
		return 0, ""
	}

	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, ""
	}

	return id, param
}

func main() {

	flag.Parse()

	var baseURL string

	if inputURL != "" {
		baseURL = strings.TrimSpace(inputURL)
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter target URL: ")
		input, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(input)
	}

	fmt.Println("[+] Base URL:", baseURL)

	fmt.Println("[*] Discovering endpoints...")
	endpoints := discoverEndpoints(baseURL)

	file, _ := os.Create("endpoints.txt")
	defer file.Close()

	var results []Result

	for _, ep := range endpoints {

		ep = strings.ReplaceAll(ep, "\n", "")
		ep = strings.ReplaceAll(ep, "\r", "")
		ep = strings.TrimSpace(ep)

		if ep == "" {
			continue
		}

		// Ensure leading slash
		if !strings.HasPrefix(ep, "/") {
			ep = "/" + ep
		}

		fullURL := strings.TrimRight(baseURL, "/") + ep

		fmt.Println("[+] Testing:", fullURL)

		file.WriteString(fullURL + "\n")

		i, err := strconv.Atoi(idorParam)
		if err != nil {
			i = 0
		}
		if res := testIDOR(fullURL, i, authHeader); res != "" {
			results = append(results, Result{fullURL, res})
		}
		if res := testAuth(fullURL); res != "" {
			results = append(results, Result{fullURL, res})
		}
		if res := testRateLimit(fullURL); res != "" {
			results = append(results, Result{fullURL, res})
		}
	}

	fmt.Println("[*] Running Nuclei...")
	// runNuclei("endpoints.txt")

	saveReport(results)
	fmt.Println("[+] Report saved to report.json")
}

package main

import (
	"api-scanner/endpoint"
	"api-scanner/nuclei"
	"api-scanner/report"
	"api-scanner/security"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func promptVulnType() int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Select vulnerability type to test:")
		fmt.Println("1) IDOR")
		fmt.Println("2) JWT")
		fmt.Print("Enter 1 or 2: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Input error: %v\n", err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)
		if input == "1" || input == "2" {
			value, _ := strconv.Atoi(input)
			return value
		}
		fmt.Println("Invalid choice. Please enter 1 for IDOR or 2 for JWT.")
	}
}

func promptNucleiTest() bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Do you want to run Nuclei tests? (y/n): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Input error: %v\n", err)
			return false
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Invalid choice. Please enter 'y' for yes or 'n' for no.")
	}
}

func main() {
	apiFile := flag.String("api", "endpoints_and_passwords/API.json", "path to the OpenAPI JSON input file")
	customURL := flag.String("url", "", "custom base URL to use instead of the API JSON server URL")
	vulnType := flag.Int("type", 0, "vulnerability type to test: 1 for IDOR, 2 for JWT")
	flag.Parse()

	if *vulnType == 0 {
		*vulnType = promptVulnType()
	}

	url, paths, err := endpoint.ParseURLAndPaths(*apiFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse API file %q: %v\n", *apiFile, err)
		os.Exit(1)
	}

	if *customURL != "" {
		url = *customURL
	}

	if url == "" {
		fmt.Fprintf(os.Stderr, "No base URL found in %q and no -url supplied\n", *apiFile)
		os.Exit(2)
	}

	println("Using base URL:", url)
	println("Selected vulnerability type:", *vulnType)

	var issues []string
	var results []report.Result

	switch *vulnType {
	case 1:
		fmt.Println("Running IDOR vulnerability test...")
		authToken, authType, authFound := security.GetAuthToken(url, paths)
		authHeader := ""
		if authFound {
			authHeader = strings.TrimSpace(authType + " " + authToken)
			fmt.Println("Login detected; testing IDOR with auth token.")
		} else {
			fmt.Println("No auth token found; testing IDOR without auth.")
		}
		issues = security.TestIDORPaths(url, paths, authHeader)
		if len(issues) > 0 {
			fmt.Println("Found IDOR issues:")
			fmt.Printf("Count: %d\n", len(issues)/2)
			for iterator := 0; iterator+1 < len(issues); iterator += 2 {
				results = append(results, report.Result{Issue: issues[iterator], Endpoint: issues[iterator+1]})
			}
		} else {
			fmt.Println("No IDOR issues found.")
		}
	case 2:
		fmt.Println("Running JWT vulnerability test...")
		loginURL, credentials := security.TestLoginIsJWT(url, paths)
		if loginURL == "" {
			println(" No login is detected.....! ")
			os.Exit(3)
		}

		fmt.Println("login url and crendtails{ email: ", credentials.Email, ", password :", credentials.Password, "}")
		issues, err = security.TestJWT(loginURL, paths, credentials)
		if err != nil {
			fmt.Println("Error:", err)
		}

		if len(issues) > 0 {
			fmt.Println("Found JWT issues:")
			fmt.Printf("Count: %d\n", len(issues)/3)
			for iterator := 0; iterator+2 < len(issues); iterator += 3 {
				results = append(results, report.Result{Issue: issues[iterator], Endpoint: issues[iterator+1], Token: issues[iterator+2]})
			}
		} else {
			fmt.Println("No JWT issues found.")
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown vuln type %d. Use 1 for IDOR or 2 for JWT\n", *vulnType)
		os.Exit(1)
	}
	report.SaveReport(results)
	if len(results) > 0 {
		fmt.Println("Results saved to report.json")
	} else {
		fmt.Println("No issues to save; report.json may be empty.")
	}
	fmt.Println("Vulnerability test finished.")

	if promptNucleiTest() {
		fmt.Println("Running Nuclei tests...")
		nuclei.RunNuclei(url, paths, results)
	} else {
		fmt.Println("Nuclei tests skipped.")
	}
}

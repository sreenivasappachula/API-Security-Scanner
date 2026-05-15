package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	filename := "API.json" //filename from local

	url, paths := getURLandPATH(filename)

	// printing url and paths
	println(url)
	for _, path := range paths {
		println(path)
	}

	url, credentials := testLogin_is_JWT(url, paths)

	if url == "" {
		println(" No login is detected.....! ")
		os.Exit(3)
	}

	fmt.Println("login url and crendtails{ email: ", credentials.email, ", password :", credentials.password, "}")
	issues, err := testJWT(url, paths, credentials)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// write the issues to report.json
	var results []Result

	if len(issues) > 0 {
		fmt.Println("Found Issues:")
		//get the isssue length print the issue and the endpoint, and the token
		fmt.Printf("Length : %d\n", len(issues))
		for iterator := 0; iterator < len(issues); iterator++ {
			results = append(results, Result{Issue: issues[iterator], Endpoint: issues[iterator+1], Token: issues[iterator+2]})
			iterator = iterator + 2
		}
	}
	saveReport(results)
	//pass the Paths and check for cves from nuclei
	runNuclei(url, paths, results)

}

func getURLandPATH(filename string) (string, []string) {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error:", err)
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

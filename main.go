package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter target URL: ")
	baseURL, _ := reader.ReadString('\n')

	baseURL = baseURL[:len(baseURL)-1]

	fmt.Println("[*] Discovering endpoints...")
	endpoints := discoverEndpoints(baseURL)

	file, _ := os.Create("endpoints.txt")
	defer file.Close()

	var results []Result

	for _, ep := range endpoints {
		fullURL := baseURL + ep
		fmt.Println("[+] Testing:", fullURL)

		file.WriteString(fullURL + "\n")

		if res := testIDOR(baseURL, ep); res != "" {
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
	runNuclei("endpoints.txt")

	saveReport(results)
	fmt.Println("[+] Report saved to report.json")
}
package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func runNuclei(targetURL string, paths []string, token []Result) {

	for i := 0; i < len(paths); i++ {
		for iterator := 0; iterator < len(token); iterator++ {
			if token[iterator].Token == "" {
				//fmt.Println("Using token for Nuclei:", token[iterator].Token)
				continue
			}

			parsedURL, err := url.Parse(targetURL)
			if err != nil {
				fmt.Println("invalid URL:", err)
				return
			}
			baseURL := parsedURL.Scheme + "://" + parsedURL.Host
			//replace localhost with 127.0.0.1
			baseURL = strings.Replace(baseURL, "localhost", "127.0.0.1", 1)
			fullURL := strings.TrimRight(baseURL, "") + "/" + strings.TrimLeft(paths[i], "/")
			fmt.Println("Running Nuclei for endpoint:", fullURL)
			cmd := exec.Command(
				"nuclei",
				"-u", fullURL,
				"-tags", "cve,tech,exposure,misconfig",
				"-o", "nuclei-report.json",
				"-H", "Authorization: Bearer "+token[iterator].Token,
			)
			i++
			output, err := cmd.CombinedOutput()

			if err != nil {
				fmt.Println("Nuclei error:", err)
			}

			fmt.Println(string(output))
		}
	}
}

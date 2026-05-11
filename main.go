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

	resp := testLogin_is_JWT(url, paths)
	println(resp)
	//testJWT(url, paths)

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

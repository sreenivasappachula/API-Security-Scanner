package main

import (
	"api-scanner/endpoint"
	"api-scanner/nuclei"
	"api-scanner/report"
	"api-scanner/security"
	"fmt"
	"os"
)

func main() {

	filename := "API.json" //filename from local

	url, paths := endpoint.ParseURLAndPaths(filename)

	// printing url and paths
	println(url)
	for _, path := range paths {
		println(path)
	}

	url, credentials := security.TestLoginIsJWT(url, paths)

	if url == "" {
		println(" No login is detected.....! ")
		os.Exit(3)
	}

	fmt.Println("login url and crendtails{ email: ", credentials.Email, ", password :", credentials.Password, "}")
	issues, err := security.TestJWT(url, paths, credentials)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// write the issues to report.json
	var results []report.Result

	if len(issues) > 0 {
		fmt.Println("Found Issues:")
		//get the isssue length print the issue and the endpoint, and the token
		fmt.Printf("Length : %d\n", len(issues))
		for iterator := 0; iterator < len(issues); iterator++ {
			results = append(results, report.Result{Issue: issues[iterator], Endpoint: issues[iterator+1], Token: issues[iterator+2]})
			iterator = iterator + 2
		}
	}
	report.SaveReport(results)
	//pass the Paths and check for cves from nuclei
	nuclei.RunNuclei(url, paths, results)

}

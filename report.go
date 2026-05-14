package main

import (
	"encoding/json"
	"os"
)

type Result struct {
	Endpoint string `json:"endpoint"`
	Issue    string `json:"issue"`
	Token    string `json:"token,omitempty"`
}

func saveReport(results []Result) {
	file, _ := os.Create("report.json")
	defer file.Close()

	json.NewEncoder(file).Encode(results)
}

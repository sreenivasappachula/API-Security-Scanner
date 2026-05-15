package main

import (
	"encoding/json"
	"os"
)

type Result struct {
	Issue    string `json:"issue"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token,omitempty"`
}

func saveReport(results []Result) {
	file, _ := os.Create("report.json")
	defer file.Close()

	json.NewEncoder(file).Encode(results)
}

package report

import (
	"encoding/json"
	"os"
)

type Result struct {
	Issue    string `json:"issue"`
	Endpoint string `json:"endpoint"`
	Token    string `json:"token,omitempty"`
}

func SaveReport(results []Result) {
	file, _ := os.Create("report.json")
	defer file.Close()

	json.NewEncoder(file).Encode(results)
}

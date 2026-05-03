package main

import (
	"os/exec"
)

func runNuclei(file string) {
	cmd := exec.Command("nuclei", "-l", file, "-o", "nuclei_output.txt")
	cmd.Run()
}

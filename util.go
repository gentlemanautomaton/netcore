package main

import (
	"os/exec"
	"strings"
)

func getHostname() (string, error) {
	fqdn, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(fqdn)), nil
}

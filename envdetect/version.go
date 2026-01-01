package envdetect

import (
	"fmt"
	"os/exec"
	"strings"
)

// FindVersion finds the Go installation version and returns it as a semantic version (semver) string.
// Returns the version string and any errors.
// Requires Go to be installed and accessible in the system PATH.
func FindVersion() (string, error) {
	version, err := exec.Command("go", "version").Output() // Finds the version of Go, but as a string with other elements

	if err != nil {
		return "", fmt.Errorf("there was an error in finding your Go version: %w. \n Have you installed Go, and is it in your PATH?", err)
	}
	// Split the string, get the third element, trim the "go" prefix, and add "v".
	cleanedVersion := "v" + strings.TrimPrefix(strings.Split(string(version), " ")[2], "go")
	return cleanedVersion, nil
}

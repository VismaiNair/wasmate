package envdetect

import (
	"fmt"     // Format & Output Strings
	"os/exec" // Execute CLI commands
)

// FindGOROOT finds the Go installation directory, which is also known as the GOROOT
// Returns the string of the filepath of the GOROOT and any errors.
// Requires Go to be installed and accessible in the system PATH.
func FindGOROOT() (string, error) {
	goroot, gorootErr := exec.Command("go", "env", "GOROOT").Output() // Finds the directory of the Go installation (GOROOT)

	if gorootErr != nil {
		return "", fmt.Errorf("there was an error in finding your GOROOT: %w. \n Have you installed Go, and is it in your PATH?", gorootErr)
	}
	return string(goroot), nil
}

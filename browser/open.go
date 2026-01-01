package browser

import (
	"os/exec"
	"runtime"
)

// Open() opens the default browser through the platform's cmd.
// It returns an error if the command fails to start.
// To be used by the wasmate run command to open the served files in the browser.
func Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", "netbsd", etc.
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

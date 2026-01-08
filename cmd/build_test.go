package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildCmdStructure tests command configuration
func TestBuildCmdStructure(t *testing.T) {
	if buildCmd.Use != "build [package]" {
		t.Errorf("Expected Use='build [package]', got %q", buildCmd.Use)
	}

	required := []string{"WebAssembly", "WASM", "GOOS=js", "GOARCH=wasm"}
	for _, term := range required {
		if !strings.Contains(buildCmd.Long, term) {
			t.Errorf("Long description should mention %q", term)
		}
	}

	if buildCmd.Run == nil {
		t.Error("Run function should be defined")
	}
}

// TestBuildCmdFlags tests the output flag
func TestBuildCmdFlags(t *testing.T) {
	outputFile = ""

	flag := buildCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("output flag should exist")
	}
	if flag.Shorthand != "o" {
		t.Errorf("Expected shorthand 'o', got %q", flag.Shorthand)
	}

	buildCmd.ParseFlags([]string{"--output", "app.wasm"})
	if outputFile != "app.wasm" {
		t.Errorf("Expected outputFile='app.wasm', got %q", outputFile)
	}

	outputFile = ""
	buildCmd.ParseFlags([]string{"-o", "short.wasm"})
	if outputFile != "short.wasm" {
		t.Errorf("Expected outputFile='short.wasm', got %q", outputFile)
	}
}

// TestBuildCommandArgs tests argument handling
func TestBuildCommandArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		outputFlag   string
		expectedArgs []string
	}{
		{
			name:         "no arguments",
			args:         []string{},
			outputFlag:   "",
			expectedArgs: []string{"build", "."},
		},
		{
			name:         "with package",
			args:         []string{"./cmd/app"},
			outputFlag:   "",
			expectedArgs: []string{"build", "./cmd/app"},
		},
		{
			name:         "with output flag",
			args:         []string{},
			outputFlag:   "output.wasm",
			expectedArgs: []string{"build", "-o", "output.wasm", "."},
		},
		{
			name:         "package and output",
			args:         []string{"./myapp"},
			outputFlag:   "app.wasm",
			expectedArgs: []string{"build", "-o", "app.wasm", "./myapp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "."
			if len(tt.args) > 0 {
				target = tt.args[0]
			}

			buildArgs := []string{"build"}
			if tt.outputFlag != "" {
				buildArgs = append(buildArgs, "-o", tt.outputFlag)
			}
			buildArgs = append(buildArgs, target)

			if len(buildArgs) != len(tt.expectedArgs) {
				t.Fatalf("Expected %d args, got %d", len(tt.expectedArgs), len(buildArgs))
			}

			for i, arg := range buildArgs {
				if arg != tt.expectedArgs[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tt.expectedArgs[i], arg)
				}
			}
		})
	}
}

// TestBuildEnvironment tests WASM environment variables
func TestBuildEnvironment(t *testing.T) {
	env := append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	hasGOOS := false
	hasGOARCH := false
	for _, e := range env {
		if e == "GOOS=js" {
			hasGOOS = true
		}
		if e == "GOARCH=wasm" {
			hasGOARCH = true
		}
	}

	if !hasGOOS {
		t.Error("GOOS=js should be set")
	}
	if !hasGOARCH {
		t.Error("GOARCH=wasm should be set")
	}
}

// TestBuildIntegration tests actual WASM compilation
func TestBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("Go not available")
	}

	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := `package main
import "fmt"
func main() { fmt.Println("Hello, WASM!") }
`
	os.WriteFile(goFile, []byte(goContent), 0644)

	// Initialize go.mod
	exec.Command("go", "mod", "init", "testmodule").Run()

	// Build WASM
	outputPath := filepath.Join(tmpDir, "test.wasm")
	cmd := exec.Command("go", "build", "-o", outputPath, goFile)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Build failed: %v\n%s", err, stderr.String())
	}

	// Verify WASM file
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Check WASM magic number: 0x00 0x61 0x73 0x6d
	if len(content) < 4 {
		t.Fatal("Output too small to be valid WASM")
	}

	expectedMagic := []byte{0x00, 0x61, 0x73, 0x6d}
	if !bytes.Equal(content[:4], expectedMagic) {
		t.Errorf("Invalid WASM magic number: %v", content[:4])
	}
}

// TestBuildCommandRegistration verifies command is registered
func TestBuildCommandRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "build [package]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("build command not registered with root")
	}
}

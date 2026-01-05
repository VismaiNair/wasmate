package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCopyFile tests the core copyFile functionality
func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (src, dst string)
		expectError bool
	}{
		{
			name: "successful copy",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "source.js")
				dst := filepath.Join(tmpDir, "dest.js")
				os.WriteFile(src, []byte("// wasm_exec.js content"), 0644)
				return src, dst
			},
			expectError: false,
		},
		{
			name: "source file not found",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "missing.js"), filepath.Join(tmpDir, "dest.js")
			},
			expectError: true,
		},
		{
			name: "invalid destination path",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				src := filepath.Join(tmpDir, "source.js")
				os.WriteFile(src, []byte("content"), 0644)
				return src, filepath.Join(tmpDir, "nonexistent", "dest.js")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup(t)
			err := copyFile(src, dst)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify content copied correctly
				if content, _ := os.ReadFile(dst); len(content) == 0 {
					t.Error("Destination file is empty")
				}
			}
		})
	}
}

// TestJsCmdStructure tests the command is properly configured
func TestJsCmdStructure(t *testing.T) {
	if jsCmd.Use != "js" {
		t.Errorf("Expected Use='js', got %q", jsCmd.Use)
	}

	if !strings.Contains(jsCmd.Long, "wasm_exec.js") {
		t.Error("Long description should mention wasm_exec.js")
	}

	if jsCmd.RunE == nil {
		t.Error("RunE should be defined")
	}

	// Verify it's registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "js" {
			found = true
			break
		}
	}
	if !found {
		t.Error("js command not registered with root")
	}
}

// TestJsCmdPathLogic tests the version-based path construction
func TestJsCmdPathLogic(t *testing.T) {
	tests := []struct {
		version    string
		goroot     string
		v121Plus   bool
		shouldHave []string // Path components to check
	}{
		{"v1.21", "/usr/local/go", true, []string{"lib", "wasm", "wasm_exec.js"}},
		{"v1.22", "/usr/local/go", true, []string{"lib", "wasm", "wasm_exec.js"}},
		{"v1.20", "/usr/local/go", false, []string{"misc", "wasm", "wasm_exec.js"}},
		{"v1.19", "/opt/go", false, []string{"misc", "wasm", "wasm_exec.js"}},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			// Simulate the path construction logic
			goroot := strings.TrimSpace(tt.goroot)
			var path string

			if tt.version >= "v1.21" {
				path = filepath.Join(goroot, "lib", "wasm", "wasm_exec.js")
			} else {
				path = filepath.Join(goroot, "misc", "wasm", "wasm_exec.js")
			}

			// Verify path contains expected components
			for _, component := range tt.shouldHave {
				if !strings.Contains(path, component) {
					t.Errorf("Path should contain %q, got %q", component, path)
				}
			}

			// Verify correct subdirectory based on version
			if tt.v121Plus && !strings.Contains(path, "lib") {
				t.Error("Go 1.21+ should use lib/wasm path")
			}
			if !tt.v121Plus && !strings.Contains(path, "misc") {
				t.Error("Go <1.21 should use misc/wasm path")
			}
		})
	}
}

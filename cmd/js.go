package cmd

import (
	"fmt"
	"io"
	"os"            // Execute CLI commands
	"path/filepath" // Work with file paths
	"strings"       // Deal with strings

	"github.com/spf13/cobra"                  // Importing the Cobra library for CLI apps
	"github.com/vismainair/wasmate/envdetect" // envdetect (part of wasmate) for go environment detection
	"golang.org/x/mod/semver"                 // The semver library by Golang for comparing different versions
)

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("could not create destination file %s: %w", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("could not copy file: %w", err)
	}
	return nil
}

// jsCmd represents the js command
var jsCmd = &cobra.Command{
	Use:   "js",
	Short: "Copies wasm_exec.js from Go installation so that WASM can be executed.",
	Long:  `The command wasmate js copies wasm_exec.js from your go installation. This is required for running WASM modules in the browser if they were compiled from Go or Node.js on your machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		goroot, gorootErr := envdetect.FindGOROOT()
		if gorootErr != nil {
			return gorootErr
		}

		version, versionErr := envdetect.FindVersion()
		if versionErr != nil {
			return versionErr
		}

		var src string
		if semver.Compare(version, "v1.21") >= 0 {
			src = filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js")
			fmt.Println("Go is greater than v1.21.")
		} else if semver.Compare(version, "v1.21") == -1 {
			src = filepath.Join(strings.TrimSpace(string(goroot)), "misc", "wasm", "wasm_exec.js")
		}

		if err := copyFile(src, "wasm_exec.js"); err != nil {
			return fmt.Errorf("there was an error with copying wasm_exec.js: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(jsCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

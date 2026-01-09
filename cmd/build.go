package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var outputFile string

var buildCmd = &cobra.Command{
	Use:   "build [package]",
	Short: "Build Go code to WebAssembly.",
	Long:  `Build Go code to WebAssembly (WASM) format. /nYou can specify a package or go file to build. If neither are specified, it defaults to all go files in the current directory. /nUnder the hood, it runs the command 'go build GOOS=js GOARCH=wasm'. This command is useful for preparing Go applications to run in a web environment using WebAssembly.`,
	Run: func(cmd *cobra.Command, args []string) {
		target := "." // Default target is current directory
		if len(args) > 0 {
			target = args[0] // If a package is specified, use it as target
		}

		// Build the command arguments
		buildArgs := []string{"build"}

		// Add output flag if specified
		if outputFile != "" {
			buildArgs = append(buildArgs, "-o", outputFile)
		}

		// Add the target
		buildArgs = append(buildArgs, target)

		buildCmd := exec.Command("go", buildArgs...)
		buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

		// Create and start spinner
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Building WASM binary..."
		s.Start()

		// Capture output but don't display it yet (so spinner stays clean)
		output, err := buildCmd.CombinedOutput()

		// Stop spinner before showing results
		s.Stop()

		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Build failed: %v\n", err)
			// Show the build output if there was an error
			if len(output) > 0 {
				fmt.Fprintf(os.Stderr, "%s\n", output)
			}
			os.Exit(1)
		}

		// Success message
		outputName := outputFile
		if outputName == "" {
			// Determine the default output name
			if target == "." {
				// Get current directory name
				dir, _ := os.Getwd()
				outputName = dir + ".wasm"
			} else {
				outputName = target + ".wasm"
			}
		}

		fmt.Fprintf(os.Stdout, "✓ Built %s\n", outputName)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file name (e.g., main.wasm)")
	rootCmd.AddCommand(buildCmd)
}

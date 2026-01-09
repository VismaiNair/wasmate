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

// RunBuild is a shared function that can be called by other commands
// showSpinner controls whether to display the spinner animation
func RunBuild(target, output string, showSpinner bool) error {
	buildArgs := []string{"build"}

	if output != "" {
		buildArgs = append(buildArgs, "-o", output)
	}

	buildArgs = append(buildArgs, target)

	buildCmd := exec.Command("go", buildArgs...)
	buildCmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	var s *spinner.Spinner
	if showSpinner {
		s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Building WASM binary..."
		s.Start()
	}

	// Changed variable name from 'output' to 'cmdOutput'
	cmdOutput, err := buildCmd.CombinedOutput()

	if showSpinner && s != nil {
		s.Stop()
	}

	if err != nil {
		return fmt.Errorf("%w\n%s", err, cmdOutput)
	}

	return nil
}

var buildCmd = &cobra.Command{
	Use:   "build [package]",
	Short: "Build Go code to WebAssembly.",
	Long:  `Build Go code to WebAssembly (WASM) format. /nYou can specify a package or go file to build. If neither are specified, it defaults to all go files in the current directory. /nUnder the hood, it runs the command 'go build GOOS=js GOARCH=wasm'. This command is useful for preparing Go applications to run in a web environment using WebAssembly.`,
	Run: func(cmd *cobra.Command, args []string) {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		// Call the shared build function with spinner enabled
		err := RunBuild(target, outputFile, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Build failed: %v\n", err)
			os.Exit(1)
		}

		// Determine output name for success message
		outputName := outputFile
		if outputName == "" {
			// Default output name logic
			outputName = "main.wasm"
		}

		fmt.Fprintf(os.Stdout, "✓ Built %s\n", outputName)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file name (e.g., main.wasm)")
	rootCmd.AddCommand(buildCmd)
}

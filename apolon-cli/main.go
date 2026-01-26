package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkeresman01/apolon/apolon-cli/generator"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	inputDir  string
	outputDir string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "apolon",
	Short: "Type-safe ORM CLI tool",
	Long: `apolon - Type-safe ORM CLI tool

Generate field constants for your Go models to enable type-safe queries.

For use with go:generate, add this comment to your models file:
  //go:generate apolon generate`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate field constants for models",
	Long: `Generate field constants for models in the specified directory.

This command parses Go files containing structs with db tags and generates
corresponding *_fields.go files with typed field accessors for queries.`,
	Example: `  apolon generate
  apolon generate --input ./models --output ./models
  apolon generate -i ./internal/domain`,
	RunE: runGenerate,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("apolon version %s\n", version)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&inputDir, "input", "i", ".", "Input directory containing model files")
	generateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated files (default: same as input)")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Resolve input directory to absolute path
	absInput, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("error resolving input path: %w", err)
	}

	// Default output to input directory
	absOutput := absInput
	if outputDir != "" {
		absOutput, err = filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("error resolving output path: %w", err)
		}
	}

	// Verify input directory exists
	info, err := os.Stat(absInput)
	if err != nil {
		return fmt.Errorf("input directory does not exist: %s", absInput)
	}
	if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", absInput)
	}

	// Create output directory if needed
	if err := os.MkdirAll(absOutput, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	fmt.Printf("Generating field constants...\n")
	fmt.Printf("  Input:  %s\n", absInput)
	fmt.Printf("  Output: %s\n", absOutput)

	parser := generator.NewParser(absInput)
	models, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	codegen := generator.NewCodeGen(absOutput)
	for sourceFile, modelList := range models {
		if err := codegen.Generate(sourceFile, modelList); err != nil {
			return err
		}
	}

	fmt.Println("Done!")
	return nil
}

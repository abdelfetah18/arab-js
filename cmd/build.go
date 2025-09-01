package cmd

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/printer"
	"arab_js/internal/transformer"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [file_path]",
	Short: "Build using the specified file path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		fmt.Println("Building from file:", filePath)

		source, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		lexer := compiler.NewLexer(string(source))
		parser := compiler.NewParser(lexer, false)
		program := parser.Parse()
		binder.NewBinder(program).Bind()
		transformer.NewTransformer(program).Transform()

		p := printer.NewPrinter()
		p.Write(program)

		output := p.Writer.Output

		outputDir := "build"
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create build directory: %w", err)
		}

		outputFile := filepath.Join(outputDir, "index.js")
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Println("Build successful! Output written to", outputFile)
		return nil
	},
}

package cmd

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/fileloader"
	"arab_js/internal/compiler/printer"
	"arab_js/internal/package_definition"
	"arab_js/internal/transformer"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func ListFilesWithExt(root, ext string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // stop if there's a problem accessing the path
		}
		if !d.IsDir() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

var buildCmd = &cobra.Command{
	Use:   "build [project_path]",
	Short: "Build project in the specified path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath := args[0]
		fmt.Println("Building project in path:", projectPath)

		info, err := os.Stat(projectPath)
		if err != nil {
			panic(err)
		}

		if !info.IsDir() {
			panic("Expeced a Directory but got a file")
		}

		packageDefinition, err := package_definition.FromYAMLFile(filepath.Join(projectPath, "رزمة.تعريف"))
		if err != nil {
			panic(err)
		}

		files, err := ListFilesWithExt(projectPath, ".كود")
		if err != nil {
			panic(err)
		}

		fileLoader := fileloader.NewFileLoader(files)
		fileLoader.LoadSourceFiles()

		program := compiler.NewProgram(fileLoader.SourceFiles)

		_checker := checker.NewChecker(program)
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			for _, diagnostic := range _checker.Diagnostics {
				println(diagnostic.Message)
			}
			panic("Checker Errors")
		}

		transformer.NewTransformer(program, _checker.NameResolver).Transform()

		outputDir := filepath.Join(projectPath, "البناء")
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create build directory: %w", err)
		}

		for _, sourceFile := range program.SourceFiles {
			outputFileName := fmt.Sprintf("%s.js", sourceFile.Name)
			if packageDefinition.Main == sourceFile.Name {
				outputFileName = "index.js"
			}

			p := printer.NewPrinter()
			p.Write(sourceFile)
			output := p.Writer.Output

			outputFile := filepath.Join(outputDir, outputFileName)
			if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
		}

		fmt.Println("Build successful!")
		return nil
	},
}

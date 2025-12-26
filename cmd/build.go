package cmd

import (
	"arab_js/internal/compiler"
	"arab_js/internal/package_definition"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
		outputPath, _ := cmd.Flags().GetString("output")

		info, err := os.Stat(projectPath)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return buildFile(projectPath, outputPath)
		}

		return buildProject(projectPath, outputPath)
	},
}

func buildProject(projectPath string, outputPath string) error {
	fmt.Printf("Build project '%s'\n", projectPath)
	packageDefinition, err := package_definition.FromYAMLFile(filepath.Join(projectPath, "رزمة.تعريف"))
	if err != nil {
		return err
	}

	files, err := ListFilesWithExt(projectPath, ".كود")
	if err != nil {
		return err
	}

	program := compiler.NewProgram()
	program.ProgramOptions.Main = packageDefinition.Main
	program.ParseSourceFiles(files)
	program.TransformSourceFiles()

	if len(program.Diagnostics) > 0 {
		for _, diagnostic := range program.Diagnostics {
			println(diagnostic.Message)
		}
		return fmt.Errorf("checker errors")
	}

	pathInfo, err := os.Stat(outputPath)
	if err != nil || !pathInfo.IsDir() {
		if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create build directory: %w", err)
		}

	}

	err = program.WriteSourceFiles(outputPath)
	if err != nil {
		return err
	}

	return nil
}

func buildFile(filePath string, outputPath string) error {
	fmt.Printf("Build file '%s'\n", filePath)

	program := compiler.NewProgram()
	program.ProgramOptions.Main = filepath.Base(filePath)
	program.ParseSourceFiles([]string{filePath})
	program.TransformSourceFiles()

	if len(program.Diagnostics) > 0 {
		for _, diagnostic := range program.Diagnostics {
			println(diagnostic.Message)
		}
		return fmt.Errorf("checker errors")
	}

	output := program.EmitSourceFile(filePath)
	outputFileName := strings.Replace(filepath.Base(filePath), ".arts", ".js", 1)
	outputFile := filepath.Join(outputPath, outputFileName)

	if err := os.WriteFile(outputFile, []byte(output), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func init() {
	buildCmd.Flags().String("output", ".", "output path")
}

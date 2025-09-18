package cmd

import (
	"arab_js/internal/compiler"
	"arab_js/internal/package_definition"
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

		program := compiler.NewProgram()
		program.ProgramOptions.Main = packageDefinition.Main
		program.ParseSourceFiles(files)
		program.TransformSourceFiles()

		if len(program.Diagnostics) > 0 {
			for _, diagnostic := range program.Diagnostics {
				println(diagnostic.Message)
			}
			panic("Checker Errors")
		}

		outputDir := filepath.Join(projectPath, "البناء")
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create build directory: %w", err)
		}

		err = program.WriteSourceFiles(outputDir)
		if err != nil {
			panic(err)
		}

		fmt.Println("Build successful!")
		return nil
	},
}

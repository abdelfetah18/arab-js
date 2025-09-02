package cmd

import (
	"arab_js/internal/package_definition"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [project_name]",
	Short: "Create a new project with the specified name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		fmt.Println("Create a new project called ", projectName)
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		projectPath := filepath.Join(cwd, projectName)
		err = os.Mkdir(projectPath, 0755)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(
			filepath.Join(projectPath, "رزمة.تعريف"),
			[]byte(package_definition.DefaultPackageDefinitionYAML(projectName)),
			0644,
		)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(
			filepath.Join(projectPath, "الرئيسي.كود"),
			[]byte{},
			0644,
		)
		if err != nil {
			panic(err)
		}

		return nil
	},
}

package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arab-js",
	Short: "Arab JS tool",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(lspCmd)
	rootCmd.AddCommand(apiCmd)
}

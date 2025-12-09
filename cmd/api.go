package cmd

import (
	"arab_js/internal/api"

	"github.com/spf13/cobra"
)

var (
	apiLogPath string
	apiCmd     = &cobra.Command{
		Use:   "api",
		Short: "Start API Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			initLogger(apiLogPath)
			api.StartServer()
			return nil
		},
	}
)

func init() {
	apiCmd.Flags().StringVar(&apiLogPath, "logs", "", "logs file path")
}

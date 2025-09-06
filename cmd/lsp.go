package cmd

import (
	"arab_js/internal/lsp"
	"fmt"
	"log"
	"os"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/spf13/cobra"
)

var (
	logPath string
	lspCmd  = &cobra.Command{
		Use:   "lsp",
		Short: "Start Language Server Protocol (LSP) Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			initLogger(logPath)
			lsp.StartLSP()
			return nil
		},
	}
)

func init() {
	lspCmd.Flags().StringVar(&logPath, "logs", "", "logs file path")
}

func initLogger(path string) {
	var logger *log.Logger

	if path == "" {
		// fallback to stderr
		logger = log.New(os.Stderr, "", log.LstdFlags)
		logs.Init(logger)
		return
	}

	// Open or create log file in append mode
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file %s: %v", path, err))
	}

	logger = log.New(f, "", log.LstdFlags)
	logs.Init(logger)
}

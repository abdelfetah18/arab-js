package lsp

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func StartLSP() {
	server := lsp.NewServer(&lsp.Options{
		CompletionProvider: &defines.CompletionOptions{
			TriggerCharacters: &[]string{"."},
		},
	})

	handlers := NewHandlers(server)

	server.OnInitialized(func(ctx context.Context, req *defines.InitializeParams) (err error) {
		logs.Println("OnInitialized")
		return nil
	})

	server.OnDidOpenTextDocument(handlers.OnDidOpenTextDocumentHandler)
	server.OnDidChangeTextDocument(handlers.OnDidChangeTextDocument)
	server.OnCompletion(handlers.OnCompletionHandler)

	for _, m := range server.GetMethods() {
		if m != nil {
			logs.Printf("m.Name=%s\n", m.Name)
		}
	}
	server.Run()
}

func ReadFile(filename defines.DocumentUri) ([]string, error) {
	enEscapeUrl, _ := url.QueryUnescape(string(filename))
	data, err := os.ReadFile(enEscapeUrl[6:])
	if err != nil {
		return nil, err
	}
	content := string(data)
	line := strings.Split(content, "\n")
	return line, nil
}

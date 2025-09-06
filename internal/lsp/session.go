package lsp

import "arab_js/internal/compiler"

type Session struct {
	FileLoader compiler.FileLoader
}

func NewSession(fileLoader compiler.FileLoader) *Session {
	return &Session{
		FileLoader: fileLoader,
	}
}

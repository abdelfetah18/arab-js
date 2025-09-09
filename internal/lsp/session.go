package lsp

import "arab_js/internal/compiler/fileloader"

type Session struct {
	FileLoader fileloader.FileLoader
}

func NewSession(fileLoader fileloader.FileLoader) *Session {
	return &Session{
		FileLoader: fileLoader,
	}
}

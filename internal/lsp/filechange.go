package lsp

import "github.com/TobiasYin/go-lsp/lsp/defines"

type FileChangeKind int

const (
	FileChangeKindOpen = iota
	FileChangeKindChange
)

type FileChange struct {
	uri     defines.DocumentUri
	content string
	version int
	Kind    FileChangeKind
}

func NewFileChange(uri defines.DocumentUri, content string, version int, kind FileChangeKind) *FileChange {
	return &FileChange{
		uri:     uri,
		content: content,
		version: version,
		Kind:    kind,
	}
}

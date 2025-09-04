package compiler

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
)

type FileLoader struct {
	files       []string
	SourceFiles []*ast.SourceFile
}

func NewFileLoader(files []string) *FileLoader {
	return &FileLoader{
		files:       files,
		SourceFiles: []*ast.SourceFile{},
	}
}

func (l *FileLoader) LoadSourceFiles() {
	for _, file := range l.files {
		sourceFile := GetSourceFile(file)
		binder.BindSourceFile(sourceFile)
		l.SourceFiles = append(l.SourceFiles, sourceFile)
	}
}

package fileloader

import (
	"arab_js/internal/binder"
	"arab_js/internal/bundled"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"

	"github.com/TobiasYin/go-lsp/logs"
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
	// Load Dom Library
	libFileContent := bundled.ReadLibFile(bundled.LibNameDom)
	sourceFile := compiler.ParseSourceFile(libFileContent)
	binder.BindSourceFile(sourceFile)
	l.SourceFiles = append(l.SourceFiles, sourceFile)

	for _, file := range l.files {
		sourceFile := compiler.GetSourceFile(file)
		binder.BindSourceFile(sourceFile)
		l.SourceFiles = append(l.SourceFiles, sourceFile)
	}
}

func (l *FileLoader) GetSourceFile(filename string) *ast.SourceFile {
	for _, sourceFile := range l.SourceFiles {
		logs.Printf("sourceFile.Name=%s\n", sourceFile.Name)
		if sourceFile.Name == filename {
			return sourceFile
		}
	}
	return nil
}

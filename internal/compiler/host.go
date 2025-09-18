package compiler

import (
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/parser"
	"os"
	"path/filepath"
)

func GetSourceFile(path string) *ast.SourceFile {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	sourceFile := parser.ParseSourceFile(string(data))
	sourceFile.Name = filepath.Base(path)

	return sourceFile

}

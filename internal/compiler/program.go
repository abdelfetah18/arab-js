package compiler

import (
	"arab_js/internal/compiler/ast"
)

type Program struct {
	SourceFiles []*ast.SourceFile
}

func NewProgram(sourceFiles []*ast.SourceFile) *Program {
	return &Program{
		SourceFiles: sourceFiles,
	}
}

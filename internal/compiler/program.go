package compiler

import (
	"arab_js/internal/binder"
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

func (p *Program) BindSourceFiles() {
	for _, file := range p.SourceFiles {
		binder.BindSourceFile(file)
	}
}

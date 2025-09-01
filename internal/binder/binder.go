package binder

import (
	"arab_js/internal/compiler/ast"
)

type Binder struct {
	program   *ast.Program
	container *ast.ContainerBase
}

func NewBinder(program *ast.Program) *Binder {
	return &Binder{
		program:   program,
		container: &program.ContainerBase,
	}
}

func (b *Binder) Bind() {
	b.program.Scope = &ast.Scope{
		Locals: map[string]*ast.Symbol{},
		Parent: nil,
	}

	for _, node := range b.program.Body {
		switch node.Type {
		case ast.NodeTypeVariableDeclaration:
			b.bindVariableDeclaration(node.AsVariableDeclaration())
		case ast.NodeTypeBlockStatement:
			b.bindBlockStatement(node.AsBlockStatement())
		}
	}
}

func (b *Binder) bindVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	var _type *ast.Type = nil
	if variableDeclaration.Identifier.TypeAnnotation != nil {
		_type = ast.GetTypeFromTypeAnnotationNode(variableDeclaration.Identifier.TypeAnnotation)
	}
	b.container.Scope.AddVariable(
		variableDeclaration.Identifier.Name,
		variableDeclaration.Identifier.OriginalName,
		_type,
	)
}

func (b *Binder) bindBlockStatement(blockStatement *ast.BlockStatement) {
	blockStatement.Scope.Parent = b.container.Scope
	saveContainer := b.container
	b.container = &blockStatement.ContainerBase

	for _, node := range b.program.Body {
		switch node.Type {
		case ast.NodeTypeVariableDeclaration:
			b.bindVariableDeclaration(node.AsVariableDeclaration())
		case ast.NodeTypeBlockStatement:
			b.bindBlockStatement(node.AsBlockStatement())
		}
	}

	b.container = saveContainer
}

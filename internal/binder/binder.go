package binder

import (
	"arab_js/internal/compiler/ast"
)

type Binder struct {
	sourceFile *ast.SourceFile
	container  *ast.ContainerBase
}

func NewBinder(sourceFile *ast.SourceFile) *Binder {
	return &Binder{
		sourceFile: sourceFile,
		container:  sourceFile.ContainerBaseData(),
	}
}

func BindSourceFile(sourceFile *ast.SourceFile) {
	NewBinder(sourceFile).Bind()
}

func (b *Binder) Bind() {
	b.sourceFile.Scope = &ast.Scope{
		Locals: map[string]*ast.Symbol{},
		Parent: nil,
	}

	for _, node := range b.sourceFile.Body {
		b.bindStatement(node)
	}
}

func (b *Binder) bindStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeIfStatement:
		ifStatement := node.AsIfStatement()
		b.bindStatement(ifStatement.ConsequentStatement)
		if ifStatement.AlternateStatement != nil {
			b.bindStatement(ifStatement.AlternateStatement)
		}
	case ast.NodeTypeBlockStatement:
		b.bindBlockStatement(node.AsBlockStatement())
	case ast.NodeTypeVariableDeclaration:
		b.bindVariableDeclaration(node.AsVariableDeclaration())
	case ast.NodeTypeInterfaceDeclaration:
		b.bindInterfaceDeclaration(node.AsInterfaceDeclaration())
	case ast.NodeTypeFunctionDeclaration:
		b.bindFunctionDeclaration(node.AsFunctionDeclaration())
	case ast.NodeTypeForStatement:
		b.bindForStatement(node.AsForStatement())
	}
}

func (b *Binder) bindVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Name.Type == ast.NodeTypeIdentifier {
		variableDeclaration.Symbol = b.container.Scope.AddVariable(
			variableDeclaration.Name.AsIdentifier().Name,
			variableDeclaration.Name.AsIdentifier().OriginalName,
			variableDeclaration.AsNode(),
		)
	}
}

func (b *Binder) bindBlockStatement(blockStatement *ast.BlockStatement) {
	saveContainer := b.container

	if canCreateNewScope(blockStatement.AsNode()) {
		blockStatement.Scope = &ast.Scope{}
		blockStatement.Scope.Parent = b.container.Scope
		b.container = blockStatement.ContainerBaseData()
	}

	for _, node := range blockStatement.Body {
		b.bindStatement(node)
	}

	b.container = saveContainer
}

func (b *Binder) bindInterfaceDeclaration(interfaceDeclaration *ast.InterfaceDeclaration) {
	interfaceDeclaration.Symbol = b.container.Scope.AddVariable(
		interfaceDeclaration.Id.Name,
		nil,
		interfaceDeclaration.AsNode(),
	)
}

func (b *Binder) bindFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	functionDeclaration.Symbol = b.container.Scope.AddVariable(
		functionDeclaration.ID.Name,
		functionDeclaration.ID.OriginalName,
		functionDeclaration.AsNode(),
	)

	saveContainer := b.container
	functionDeclaration.Scope = &ast.Scope{}
	functionDeclaration.Scope.Parent = b.container.Scope
	b.container = functionDeclaration.ContainerBaseData()

	for _, param := range functionDeclaration.Params {
		b.bindParam(param)
	}

	if functionDeclaration.Body != nil {
		b.bindBlockStatement(functionDeclaration.Body)
	}

	b.container = saveContainer
}

func (b *Binder) bindParam(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeParameter:
		param := node.AsParameter()
		if param.Name == nil {
			return
		}

		switch param.Name.Type {
		case ast.NodeTypeIdentifier:
			identifier := param.Name.AsIdentifier()
			b.container.Scope.AddVariable(
				identifier.Name,
				nil,
				node,
			)
		}
	}
}

func (b *Binder) bindForStatement(forStatement *ast.ForStatement) {
	saveContainer := b.container
	forStatement.Scope = &ast.Scope{}
	forStatement.Scope.Parent = b.container.Scope
	b.container = forStatement.ContainerBaseData()

	switch forStatement.Init.Type {
	case ast.NodeTypeVariableDeclaration:
		b.bindVariableDeclaration(forStatement.Init.AsVariableDeclaration())
	}

	b.bindStatement(forStatement.Body)

	b.container = saveContainer
}

func canCreateNewScope(node *ast.Node) bool {
	if node.Parent == nil {
		return true
	}

	switch node.Parent.Type {
	case ast.NodeTypeFunctionDeclaration, ast.NodeTypeForStatement:
		return false
	default:
		return true
	}
}

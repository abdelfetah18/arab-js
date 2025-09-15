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
		b.bindStatement(ifStatement.AlternateStatement)
	case ast.NodeTypeBlockStatement:
		b.bindBlockStatement(node.AsBlockStatement())
	case ast.NodeTypeVariableDeclaration:
		b.bindVariableDeclaration(node.AsVariableDeclaration())
	case ast.NodeTypeTInterfaceDeclaration:
		b.bindTInterfaceDeclaration(node.AsTInterfaceDeclaration())
	case ast.NodeTypeFunctionDeclaration:
		b.bindFunctionDeclaration(node.AsFunctionDeclaration())
	case ast.NodeTypeForStatement:
		b.bindForStatement(node.AsForStatement())
	}
}

func (b *Binder) bindVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	var _type *ast.Type = nil
	if variableDeclaration.Identifier.TypeAnnotation != nil {
		_type = b.GetTypeFromTypeAnnotationNode(variableDeclaration.Identifier.TypeAnnotation)
	}
	variableDeclaration.Symbol = b.container.Scope.AddVariable(
		variableDeclaration.Identifier.Name,
		variableDeclaration.Identifier.OriginalName,
		_type,
	)
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

func (b *Binder) bindTInterfaceDeclaration(tInterfaceDeclaration *ast.TInterfaceDeclaration) {
	tInterfaceDeclaration.Symbol = b.container.Scope.AddVariable(
		tInterfaceDeclaration.Id.Name,
		nil,
		b.GetTypeFromTypeNode(tInterfaceDeclaration.Body.AsNode()),
	)
}

func (b *Binder) GetTypeFromTypeAnnotationNode(node *ast.TTypeAnnotation) *ast.Type {
	return b.GetTypeFromTypeNode(node.TypeAnnotation)
}

func (b *Binder) GetTypeFromTypeNode(node *ast.Node) *ast.Type {
	switch node.Type {
	case ast.NodeTypeTStringKeyword:
		return ast.NewType(ast.NewStringType()).AsType()
	case ast.NodeTypeTBooleanKeyword:
		return ast.NewType(ast.NewBooleanType()).AsType()
	case ast.NodeTypeTNumberKeyword:
		return ast.NewType(ast.NewNumberType()).AsType()
	case ast.NodeTypeTNullKeyword:
		return ast.NewType(ast.NewNullType()).AsType()
	case ast.NodeTypeTTypeReference:
		tTypeReference := node.AsTTypeReference()
		symbol := b.container.Scope.GetVariableSymbol(tTypeReference.TypeName.Name)
		if symbol == nil {
			return nil
		}
		return symbol.Type
	case ast.NodeTypeTInterfaceBody:
		objectType := ast.NewType(ast.NewObjectType())
		tInterfaceBody := node.AsTInterfaceBody()
		for _, node := range tInterfaceBody.Body {
			tPropertySignature := node.AsTPropertySignature()
			objectType.AddProperty(tPropertySignature.Key.AsIdentifier().Name, &ast.PropertyType{
				Name:         tPropertySignature.Key.AsIdentifier().Name,
				OriginalName: tPropertySignature.Key.AsIdentifier().OriginalName,
				Type:         b.GetTypeFromTypeAnnotationNode(tPropertySignature.TypeAnnotation),
			})
		}
		return objectType.AsType()
	}

	return nil
}

func (b *Binder) bindFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	functionType := ast.NewType(ast.NewFunctionType())

	for _, param := range functionDeclaration.Params {
		switch param.Type {
		case ast.NodeTypeIdentifier:
			identifier := param.AsIdentifier()
			if identifier.TypeAnnotation != nil {
				functionType.AddParamType(b.GetTypeFromTypeAnnotationNode(identifier.TypeAnnotation))
			}
		case ast.NodeTypeRestElement:
			restElement := param.AsRestElement()
			if restElement.TypeAnnotation != nil {
				functionType.AddParamType(b.GetTypeFromTypeAnnotationNode(restElement.TypeAnnotation))
			}
		}
	}

	if functionDeclaration.TTypeAnnotation != nil {
		functionType.ReturnType = b.GetTypeFromTypeAnnotationNode(functionDeclaration.TTypeAnnotation)
	}

	functionDeclaration.Symbol = b.container.Scope.AddVariable(
		functionDeclaration.ID.Name,
		nil,
		functionType.AsType(),
	)

	saveContainer := b.container
	functionDeclaration.Scope = &ast.Scope{}
	functionDeclaration.Scope.Parent = b.container.Scope
	b.container = functionDeclaration.ContainerBaseData()

	for _, param := range functionDeclaration.Params {
		b.bindParam(param)
	}

	b.bindBlockStatement(functionDeclaration.Body)

	b.container = saveContainer
}

func (b *Binder) bindParam(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeRestElement:
		restElement := node.AsRestElement()
		b.container.Scope.AddVariable(
			restElement.Argument.Name,
			nil,
			b.GetTypeFromTypeAnnotationNode(restElement.TypeAnnotation),
		)
	case ast.NodeTypeIdentifier:
		identifier := node.AsIdentifier()
		b.container.Scope.AddVariable(
			identifier.Name,
			nil,
			b.GetTypeFromTypeAnnotationNode(identifier.TypeAnnotation),
		)
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

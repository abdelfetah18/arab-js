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
		switch node.Type {
		case ast.NodeTypeVariableDeclaration:
			b.bindVariableDeclaration(node.AsVariableDeclaration())
		case ast.NodeTypeBlockStatement:
			b.bindBlockStatement(node.AsBlockStatement())
		case ast.NodeTypeTInterfaceDeclaration:
			b.bindTInterfaceDeclaration(node.AsTInterfaceDeclaration())
		}
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
	blockStatement.Scope.Parent = b.container.Scope
	saveContainer := b.container
	b.container = &blockStatement.ContainerBase

	for _, node := range b.sourceFile.Body {
		switch node.Type {
		case ast.NodeTypeVariableDeclaration:
			b.bindVariableDeclaration(node.AsVariableDeclaration())
		case ast.NodeTypeBlockStatement:
			b.bindBlockStatement(node.AsBlockStatement())
		}
	}

	b.container = saveContainer
}

func (b *Binder) bindTInterfaceDeclaration(tInterfaceDeclaration *ast.TInterfaceDeclaration) {
	b.container.Scope.AddVariable(
		tInterfaceDeclaration.Id.Name,
		nil,
		b.GetTypeFromTypeNode(tInterfaceDeclaration.Body.ToNode()),
	)
}

func (b *Binder) GetTypeFromTypeAnnotationNode(node *ast.TTypeAnnotation) *ast.Type {
	return b.GetTypeFromTypeNode(node.TypeAnnotation)
}

func (b *Binder) GetTypeFromTypeNode(node *ast.Node) *ast.Type {
	switch node.Type {
	case ast.NodeTypeTStringKeyword:
		return ast.NewStringType().ToType()
	case ast.NodeTypeTBooleanKeyword:
		return ast.NewBooleanType().ToType()
	case ast.NodeTypeTNumberKeyword:
		return ast.NewNumberType().ToType()
	case ast.NodeTypeTNullKeyword:
		return ast.NewNullType().ToType()
	case ast.NodeTypeTTypeReference:
		tTypeReference := node.AsTTypeReference()
		symbol := b.container.Scope.GetVariableSymbol(tTypeReference.TypeName.Name)
		if symbol == nil {
			return nil
		}
		return symbol.Type
	case ast.NodeTypeTInterfaceBody:
		objectType := ast.NewObjectType()
		tInterfaceBody := node.AsTInterfaceBody()
		for _, node := range tInterfaceBody.Body {
			tPropertySignature := node.AsTPropertySignature()
			objectType.AddProperty(tPropertySignature.Key.AsIdentifier().Name, &ast.PropertyType{
				Name:         tPropertySignature.Key.AsIdentifier().Name,
				OriginalName: tPropertySignature.Key.AsIdentifier().OriginalName,
				Type:         b.GetTypeFromTypeAnnotationNode(tPropertySignature.TypeAnnotation),
			})
		}
		return objectType.ToType()
	}

	return nil
}

package transformer

import (
	"arab_js/internal/binder"
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
)

type Program interface {
	SourceFiles() []*ast.SourceFile
	CheckSourceFiles() *binder.NameResolver
}

type Transformer struct {
	program      Program
	NameResolver *binder.NameResolver
	TypeResolver *checker.TypeResolver

	currentScope *ast.Scope
}

func NewTransformer(program Program) *Transformer {
	nameResolver := program.CheckSourceFiles()
	return &Transformer{
		program:      program,
		NameResolver: nameResolver,
		TypeResolver: checker.NewTypeResolver(nameResolver),
	}
}

func (t *Transformer) Transform() {
	for _, sourceFile := range t.program.SourceFiles() {
		t.currentScope = sourceFile.Scope
		for _, node := range sourceFile.Body {
			switch node.Type {
			case ast.NodeTypeFunctionDeclaration:
				t.transformFunctionDeclaration(node.AsFunctionDeclaration())
				continue
			default:
				t.transformStatement(node)
				continue
			}
		}
	}
}

func (t *Transformer) transformStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeExpressionStatement:
		t.transformExpression(node.AsExpressionStatement().Expression)
	case ast.NodeTypeIfStatement:
		ifStatement := node.AsIfStatement()
		t.transformExpression(ifStatement.TestExpression)
		t.transformStatement(ifStatement.ConsequentStatement)
		t.transformStatement(ifStatement.AlternateStatement)
	case ast.NodeTypeBlockStatement:
		t.transformBlockStatement(node.AsBlockStatement())
	}
}

func (t *Transformer) transformExpression(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeCallExpression:
		t.transformCallExpression(node.AsCallExpression())
	case ast.NodeTypeIdentifier:
		identifier := node.AsIdentifier()
		symbol := t.NameResolver.Resolve(identifier.Name, node)
		if symbol.OriginalName != nil {
			identifier.Name = *symbol.OriginalName
		}
	}
}

func (t *Transformer) transformCallExpression(callExpression *ast.CallExpression) {
	switch callExpression.Callee.Type {
	case ast.NodeTypeMemberExpression:
		t.transformMemberExpression(callExpression.Callee.AsMemberExpression())
	}
}

func (t *Transformer) transformMemberExpression(memberExpression *ast.MemberExpression) {
	switch memberExpression.Object.Type {
	case ast.NodeTypeMemberExpression:
		objectType := t.TypeResolver.ResolveTypeFromNode(memberExpression.Object)
		t.transformMemberExpression(memberExpression.Object.AsMemberExpression())
		t.transformProperty(memberExpression.Property, objectType.AsObjectType())
	case ast.NodeTypeIdentifier:
		objectType := t.TypeResolver.ResolveTypeFromNode(memberExpression.Object)
		if objectType.Flags&checker.TypeFlagsObject == checker.TypeFlagsObject {
			t.transformProperty(memberExpression.Property, objectType.AsObjectType())
		}
	}
}

func (t *Transformer) transformProperty(property *ast.Node, objectType *checker.ObjectType) {
	switch property.Type {
	case ast.NodeTypeIdentifier:
		identifier := property.AsIdentifier()
		propertyType := objectType.Properties[identifier.Name]
		if propertyType.OriginalName != nil {
			identifier.Name = *propertyType.OriginalName
		}
	}
}

func (t *Transformer) transformBlockStatement(blockStatement *ast.BlockStatement) {
	for _, node := range blockStatement.Body {
		t.currentScope = blockStatement.Scope
		t.transformStatement(node)
	}
}

func (t *Transformer) transformFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	t.transformBlockStatement(functionDeclaration.Body)
}

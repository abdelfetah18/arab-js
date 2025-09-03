package transformer

import (
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
)

type Transformer struct {
	program *compiler.Program

	sourceFile *ast.SourceFile
}

func NewTransformer(program *compiler.Program) *Transformer {
	return &Transformer{
		program: program,
	}
}

func (t *Transformer) Transform() {
	for _, sourceFile := range t.program.SourceFiles {
		t.sourceFile = sourceFile
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
		symbol := t.sourceFile.Scope.GetVariableSymbol(identifier.Name)
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
		objectType := t.sourceFile.Scope.GetTypeOfNode(memberExpression.Object)
		t.transformMemberExpression(memberExpression.Object.AsMemberExpression())
		t.transformProperty(memberExpression.Property, objectType.AsObjectType())
	case ast.NodeTypeIdentifier:
		objectIdentfier := memberExpression.Object.AsIdentifier()
		symbol := t.sourceFile.Scope.GetVariableSymbol(objectIdentfier.Name)
		objectIdentfier.Name = *symbol.OriginalName
		if symbol.Type.Flags&ast.TypeFlagsObject == ast.TypeFlagsObject {
			objectType := symbol.Type.AsObjectType()
			t.transformProperty(memberExpression.Property, objectType)
		}
	}
}

func (t *Transformer) transformProperty(property *ast.Node, objectType *ast.ObjectType) {
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
		t.transformStatement(node)
	}
}

func (t *Transformer) transformFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	t.transformBlockStatement(functionDeclaration.Body)
}

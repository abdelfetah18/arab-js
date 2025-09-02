package transformer

import (
	"arab_js/internal/compiler/ast"
)

type Transformer struct {
	program *ast.Program
}

func NewTransformer(program *ast.Program) *Transformer {
	return &Transformer{
		program: program,
	}
}

func (t *Transformer) Transform() {
	for _, node := range t.program.Body {
		t.transformStatement(node)
	}
}

func (t *Transformer) transformStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeExpressionStatement:
		t.transformExpression(node.AsExpressionStatement().Expression)
	}
}

func (t *Transformer) transformExpression(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeCallExpression:
		t.transformCallExpression(node.AsCallExpression())
	case ast.NodeTypeIdentifier:
		identifier := node.AsIdentifier()
		symbol := t.program.Scope.GetVariableSymbol(identifier.Name)
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
		objectType := t.program.Scope.GetTypeOfNode(memberExpression.Object)
		t.transformMemberExpression(memberExpression.Object.AsMemberExpression())
		t.transformProperty(memberExpression.Property, objectType.AsObjectType())
	case ast.NodeTypeIdentifier:
		objectIdentfier := memberExpression.Object.AsIdentifier()
		symbol := t.program.Scope.GetVariableSymbol(objectIdentfier.Name)
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

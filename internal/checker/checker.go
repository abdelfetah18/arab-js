package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
	"fmt"
)

type Program interface {
	SourceFiles() []*ast.SourceFile
	BindSourceFiles()
}

type Checker struct {
	program      Program
	NameResolver *binder.NameResolver
	TypeResolver *TypeResolver
	Diagnostics  []*ast.Diagnostic

	currentSourceFile *ast.SourceFile
}

func NewChecker(program Program) *Checker {
	program.BindSourceFiles()

	globalScope := &ast.Scope{}

	for _, sourceFile := range program.SourceFiles() {
		if !ast.IsExternalModule(sourceFile) {
			globalScope.MergeScopeLocals(sourceFile.Scope)
		}
	}

	nameResolver := binder.NewNameResolver(globalScope)

	return &Checker{
		program:      program,
		Diagnostics:  []*ast.Diagnostic{},
		NameResolver: nameResolver,
		TypeResolver: NewTypeResolver(nameResolver),
	}
}

func (c *Checker) error(location ast.Location, message string) {
	c.Diagnostics = append(c.Diagnostics,
		ast.NewDiagnostic(
			c.currentSourceFile,
			location,
			message,
		),
	)
}

func (c *Checker) errorf(location ast.Location, format string, a ...any) {
	c.error(location, fmt.Sprintf(format, a...))
}

func (c *Checker) Check() {
	for _, sourceFile := range c.program.SourceFiles() {
		c.currentSourceFile = sourceFile
		for _, node := range sourceFile.Body {
			c.checkStatement(node)
		}
	}
}

func (c *Checker) checkStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeVariableDeclaration:
		c.checkVariableDeclaration(node.AsVariableDeclaration())
	case ast.NodeTypeExpressionStatement:
		c.checkExpression(node.AsExpressionStatement().Expression)
	}
}

func (c *Checker) checkVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Identifier.TypeAnnotation == nil {
		// NOTE: Identifier is any so we skip
		return
	}

	identifierType := c.TypeResolver.ResolveTypeNode(variableDeclaration.TypeNode())

	if variableDeclaration.Initializer == nil {
		return
	}
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)
	if initializerType == nil {
		return
	}

	if !AreTypesCompatible(identifierType, initializerType) {
		c.errorf(variableDeclaration.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.Data.Name(), initializerType.Data.Name())
		return
	}
}

func (c *Checker) checkExpression(expression *ast.Node) *Type {
	switch expression.Type {
	case ast.NodeTypeIdentifier,
		ast.NodeTypeStringLiteral,
		ast.NodeTypeDecimalLiteral,
		ast.NodeTypeBooleanLiteral,
		ast.NodeTypeNullLiteral:
		return c.TypeResolver.ResolveTypeFromNode(expression)
	case ast.NodeTypeAssignmentExpression:
		assignmentExpression := expression.AsAssignmentExpression()
		leftType := c.checkExpression(assignmentExpression.Left)
		rightType := c.checkExpression(assignmentExpression.Right)
		if leftType == nil {
			return nil
		}

		if rightType == nil {
			return nil
		}

		if !AreTypesCompatible(leftType, rightType) {
			c.errorf(assignmentExpression.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, leftType.Data.Name(), rightType.Data.Name())
			return nil
		}
		return leftType
	case ast.NodeTypeCallExpression:
		callExpression := expression.AsCallExpression()
		return c.checkCallExpression(callExpression)
	case ast.NodeTypeObjectExpression:
		objectExpression := expression.AsObjectExpression()
		return c.checkObjectExpression(objectExpression)
	case ast.NodeTypeMemberExpression:
		memberExpression := expression.AsMemberExpression()
		return c.checkMemberExpression(memberExpression)
	}

	return nil
}

func (c *Checker) checkCallExpression(callExpression *ast.CallExpression) *Type {
	_type := c.checkExpression(callExpression.Callee)
	if _type == nil {
		return nil
	}

	if _type.Flags&TypeFlagsFunction != TypeFlagsFunction {
		c.errorf(callExpression.Location, THIS_EXPRESSION_IS_NOT_CALLABLE_TYPE_0_HAS_NO_CALL_SIGNATURES, _type.Data.Name())
		return nil
	}

	functionType := _type.AsFunctionType()

	if len(functionType.Params) != len(callExpression.Args) {
		c.errorf(callExpression.Location, EXPECTED_0_ARGUMENTS_BUT_GOT_1, len(functionType.Params), len(callExpression.Args))
		return nil
	}

	for index, expression := range callExpression.Args {
		paramType := functionType.Params[index]
		_type := c.checkExpression(expression)
		if !AreTypesCompatible(paramType, _type) {
			c.errorf(callExpression.Location, ARGUMENT_OF_TYPE_0_IS_NOT_ASSIGNABLE_TO_PARAMETER_OF_TYPE_1, _type.Data.Name(), paramType.Data.Name())
			return nil
		}
	}

	return functionType.ReturnType
}

func (c *Checker) checkObjectExpression(objectExpression *ast.ObjectExpression) *Type {
	objectType := NewType(NewObjectType())
	for _, property := range objectExpression.Properties {
		switch property.Type {
		case ast.NodeTypeObjectProperty:
			objectProperty := property.AsObjectProperty()
			propertyName := objectProperty.Name()
			objectType.AddProperty(propertyName, &PropertyType{
				Name:         propertyName,
				OriginalName: nil,
				Type:         InferTypeFromNode(objectProperty.Value),
			})
		}
	}

	return objectType.AsType()
}

func (c *Checker) checkMemberExpression(memberExpression *ast.MemberExpression) *Type {
	objectType := c.checkExpression(memberExpression.Object)
	if objectType == nil {
		return nil
	}

	propertyName := memberExpression.PropertyName()

	if objectType.Flags&TypeFlagsObject != TypeFlagsObject {
		// FIXME: in JavaScript everything is object
		return nil
	}

	propertyType := objectType.AsObjectType().GetProperty(propertyName)
	if propertyType == nil {
		c.errorf(memberExpression.AsNode().Location, PROPERTY_0_DOES_NOT_EXIST_ON_TYPE_1, propertyName, objectType.Data.Name())
		return nil
	}

	return propertyType.Type
}

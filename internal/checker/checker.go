package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"fmt"
)

type Checker struct {
	program      *compiler.Program
	NameResolver *binder.NameResolver
	Diagnostics  []*ast.Diagnostic

	currentSourceFile *ast.SourceFile
}

func NewChecker(program *compiler.Program) *Checker {
	globalScope := &ast.Scope{}

	for _, sourceFile := range program.SourceFiles {
		if !ast.IsExternalModule(sourceFile) {
			globalScope.MergeScopeLocals(sourceFile.Scope)
		}
	}

	return &Checker{
		program:      program,
		Diagnostics:  []*ast.Diagnostic{},
		NameResolver: binder.NewNameResolver(globalScope),
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
	for _, sourceFile := range c.program.SourceFiles {
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

	identifierType := c.getTypeFromTypeNode(variableDeclaration.TypeNode())

	if variableDeclaration.Initializer == nil {
		return
	}
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)

	if identifierType.Data.Name() != initializerType.Data.Name() {
		c.errorf(variableDeclaration.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.Data.Name(), initializerType.Data.Name())
		return
	}
}

func (c *Checker) checkExpression(expression *ast.Node) *Type {
	switch expression.Type {
	case ast.NodeTypeIdentifier:
		identifier := expression.AsIdentifier()
		symbol := c.NameResolver.Resolve(identifier.Name, expression)
		if symbol == nil {
			c.errorf(identifier.Location, CANNOT_FIND_NAME_0, identifier.Name)
			return nil
		}
		return c.getTypeFromSymbol(symbol)
	case ast.NodeTypeStringLiteral:
		return InferTypeFromNode(expression.AsStringLiteral().AsNode())
	case ast.NodeTypeDecimalLiteral:
		return InferTypeFromNode(expression.AsDecimalLiteral().AsNode())
	case ast.NodeTypeBooleanLiteral:
		return InferTypeFromNode(expression.AsBooleanLiteral().AsNode())
	case ast.NodeTypeNullLiteral:
		return InferTypeFromNode(expression.AsNullLiteral().AsNode())
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

		if leftType.Data.Name() != rightType.Data.Name() {
			c.errorf(assignmentExpression.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, leftType.Data.Name(), rightType.Data.Name())
			return nil
		}
		return leftType
	case ast.NodeTypeCallExpression:
		callExpression := expression.AsCallExpression()
		return c.checkCallExpression(callExpression)
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
		typeName := functionType.Params[index].Data.Name()
		_type := c.checkExpression(expression)
		if typeName != _type.Data.Name() {
			c.errorf(callExpression.Location, ARGUMENT_OF_TYPE_0_IS_NOT_ASSIGNABLE_TO_PARAMETER_OF_TYPE_1, _type.Data.Name(), typeName)
			return nil
		}
	}

	return functionType.ReturnType
}

func (c *Checker) getTypeFromSymbol(symbol *ast.Symbol) *Type {
	switch symbol.Node.Type {
	case ast.NodeTypeFunctionDeclaration:
		functionDeclaration := symbol.Node.AsFunctionDeclaration()
		functionType := NewType(NewFunctionType())
		for _, param := range functionDeclaration.Params {
			functionType.AddParamType(c.getTypeFromTypeNode(param.TypeNode()))
		}
		functionType.ReturnType = c.getTypeFromTypeNode(functionDeclaration.TypeNode())
		return functionType.AsType()
	default:
		return c.getTypeFromTypeNode(symbol.Node.TypeNode())
	}
}

func (c *Checker) getTypeFromTypeNode(typeNode *ast.Node) *Type {
	if typeNode == nil {
		return nil
	}

	switch typeNode.Type {
	case ast.NodeTypeTStringKeyword:
		return NewType(NewStringType()).AsType()
	case ast.NodeTypeTBooleanKeyword:
		return NewType(NewBooleanType()).AsType()
	case ast.NodeTypeTNullKeyword:
		return NewType(NewNullType()).AsType()
	case ast.NodeTypeTNumberKeyword:
		return NewType(NewNumberType()).AsType()
	case ast.NodeTypeTTypeLiteral:
		objectType := NewObjectType()
		typeLiteral := typeNode.AsTTypeLiteral()
		for _, member := range typeLiteral.Members {
			propertySignature := member.AsTPropertySignature()
			switch propertySignature.Key.Type {
			case ast.NodeTypeIdentifier:
				identifier := propertySignature.Key.AsIdentifier()
				objectType.AddProperty(
					identifier.Name,
					&PropertyType{
						Type:         c.getTypeFromTypeNode(propertySignature.TypeNode()),
						Name:         identifier.Name,
						OriginalName: identifier.OriginalName,
					},
				)
			}
		}

		return NewType(objectType).AsType()
	}

	return nil
}

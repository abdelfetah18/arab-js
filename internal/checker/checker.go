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

	identifierType := variableDeclaration.Symbol.Type

	if variableDeclaration.Initializer == nil {
		return
	}
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)

	if identifierType.Data.Name() != initializerType.Data.Name() {
		c.errorf(variableDeclaration.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.Data.Name(), initializerType.Data.Name())
		return
	}
}

func (c *Checker) checkExpression(expression *ast.Node) *ast.Type {
	switch expression.Type {
	case ast.NodeTypeIdentifier:
		identifier := expression.AsIdentifier()
		symbol := c.NameResolver.Resolve(identifier.Name, c.currentSourceFile.Scope)
		if symbol == nil {
			c.errorf(identifier.Location, CANNOT_FIND_NAME_0, identifier.Name)
			return nil
		}

		return symbol.Type
	case ast.NodeTypeStringLiteral:
		return ast.InferTypeFromNode(expression.AsStringLiteral().AsNode())
	case ast.NodeTypeDecimalLiteral:
		return ast.InferTypeFromNode(expression.AsDecimalLiteral().AsNode())
	case ast.NodeTypeBooleanLiteral:
		return ast.InferTypeFromNode(expression.AsBooleanLiteral().AsNode())
	case ast.NodeTypeNullLiteral:
		return ast.InferTypeFromNode(expression.AsNullLiteral().AsNode())
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
	}

	return nil
}

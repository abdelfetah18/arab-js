package checker

import (
	"arab_js/internal/compiler/ast"
	"fmt"
)

type Checker struct {
	Program *ast.Program
	Errors  []string
}

func NewChecker(program *ast.Program) *Checker {
	return &Checker{
		Program: program,
		Errors:  []string{},
	}
}

func (c *Checker) error(msg string) {
	c.Errors = append(c.Errors, msg)
}

func (c *Checker) errorf(format string, a ...any) {
	c.error(fmt.Sprintf(format, a...))
}

func (c *Checker) Check() {
	for _, node := range c.Program.Body {
		c.checkStatement(node)
	}
}

func (c *Checker) checkStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeVariableDeclaration:
		c.checkVariableDeclaration(node.AsVariableDeclaration())
	}
}

func (c *Checker) checkVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Identifier.TypeAnnotation == nil {
		// NOTE: Identifier is any so we skip
		return
	}

	identifierType := variableDeclaration.Symbol.Type
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)

	if identifierType.Data.Name() != initializerType.Data.Name() {
		c.errorf(TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.Data.Name(), initializerType.Data.Name())
		return
	}
}

func (c *Checker) checkExpression(expression *ast.Node) *ast.Type {
	switch expression.Type {
	case ast.NodeTypeStringLiteral:
		return ast.InferTypeFromNode(expression.AsStringLiteral().ToNode())
	case ast.NodeTypeDecimalLiteral:
		return ast.InferTypeFromNode(expression.AsDecimalLiteral().ToNode())
	case ast.NodeTypeBooleanLiteral:
		return ast.InferTypeFromNode(expression.AsBooleanLiteral().ToNode())
	case ast.NodeTypeNullLiteral:
		return ast.InferTypeFromNode(expression.AsNullLiteral().ToNode())
	}

	return nil
}

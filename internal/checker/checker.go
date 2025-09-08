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
	Errors       []string
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
		Errors:       []string{},
		NameResolver: binder.NewNameResolver(globalScope),
	}
}

func (c *Checker) error(msg string) {
	c.Errors = append(c.Errors, msg)
}

func (c *Checker) errorf(format string, a ...any) {
	c.error(fmt.Sprintf(format, a...))
}

func (c *Checker) Check() {
	for _, sourceFile := range c.program.SourceFiles {
		for _, node := range sourceFile.Body {
			c.checkStatement(node)
		}
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

	if variableDeclaration.Initializer == nil {
		return
	}
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)

	if identifierType.Data.Name() != initializerType.Data.Name() {
		c.errorf(TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.Data.Name(), initializerType.Data.Name())
		return
	}
}

func (c *Checker) checkExpression(expression *ast.Node) *ast.Type {
	switch expression.Type {
	case ast.NodeTypeStringLiteral:
		return ast.InferTypeFromNode(expression.AsStringLiteral().AsNode())
	case ast.NodeTypeDecimalLiteral:
		return ast.InferTypeFromNode(expression.AsDecimalLiteral().AsNode())
	case ast.NodeTypeBooleanLiteral:
		return ast.InferTypeFromNode(expression.AsBooleanLiteral().AsNode())
	case ast.NodeTypeNullLiteral:
		return ast.InferTypeFromNode(expression.AsNullLiteral().AsNode())
	}

	return nil
}

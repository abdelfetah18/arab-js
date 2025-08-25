package printer

import (
	"arab_js/internal/compiler/ast"
	"fmt"
	"strings"
)

type Writer struct {
	Output string
}

func NewWriter() *Writer { return &Writer{Output: ""} }

func (writer *Writer) Write(input string)             { writer.Output += input }
func (writer *Writer) Writef(format string, a ...any) { writer.Output += fmt.Sprintf(format, a...) }

type Printer struct {
	Writer *Writer
	indent int
}

func NewPrinter() *Printer { return &Printer{Writer: NewWriter(), indent: 0} }

func (printer *Printer) increaseIndent() { printer.indent += 2 }
func (printer *Printer) decreaseIndent() { printer.indent -= 2 }

func (printer *Printer) Write(node *ast.Program) {
	printer.writeStatementList(node.Body)
}

func (printer *Printer) writeStatementList(statementList []*ast.Node) {
	for _, statement := range statementList {
		printer.writeStatement(statement)
	}
}

func (printer *Printer) writeStatement(statement *ast.Node) {
	printer.Writer.Write(strings.Repeat(" ", printer.indent))
	switch statement.Type {
	case ast.NodeTypeVariableDeclaration:
		printer.writeVariableDeclaration(statement.AsVariableDeclaration())
		printer.Writer.Write(";")
	case ast.NodeTypeIfStatement:
		printer.writeIfStatement(statement.AsIfStatement())
	case ast.NodeTypeBlockStatement:
		printer.writeBlockStatement(statement.AsBlockStatement())
	case ast.NodeTypeExpressionStatement:
		printer.writeExpressionStatement(statement.AsExpressionStatement())
	}
}

func (printer *Printer) writeExpressionStatement(expressionStatement *ast.ExpressionStatement) {
	printer.writeExpression(expressionStatement.Expression)
	printer.Writer.Write(";")
}

func (printer *Printer) writeVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	printer.Writer.Write(fmt.Sprintf("let %s = ", variableDeclaration.Identifier.Name))
	printer.writeInitializer(variableDeclaration.Initializer)
}

func (printer *Printer) writeInitializer(initializer *ast.Initializer) {
	printer.writeExpression(initializer.Expression)
}

func (printer *Printer) writeExpression(expression *ast.Node) {
	switch expression.Type {
	case ast.NodeTypeDecimalLiteral:
		printer.Writer.Write(expression.AsDecimalLiteral().Value)
	case ast.NodeTypeStringLiteral:
		printer.Writer.Writef("\"%s\"", expression.AsStringLiteral().Value)
	case ast.NodeTypeNullLiteral:
		printer.Writer.Write("null")
	case ast.NodeTypeBooleanLiteral:
		if expression.AsBooleanLiteral().Value {
			printer.Writer.Write("true")
		} else {
			printer.Writer.Write("false")
		}
	case ast.NodeTypeCallExpression:
		callExpression := expression.AsCallExpression()
		switch callExpression.Callee.Type {
		case ast.NodeTypeIdentifier:
			printer.Writer.Write(callExpression.Callee.AsIdentifier().Name)
		}
		printer.Writer.Write("(")
		for _, expression := range callExpression.Args {
			printer.writeExpression(expression)
		}
		printer.Writer.Write(")")
	}
}

func (printer *Printer) writeIfStatement(ifStatement *ast.IfStatement) {
	printer.Writer.Write("if (")
	printer.writeExpression(ifStatement.TestExpression)
	printer.Writer.Write(") ")
	printer.writeStatement(ifStatement.ConsequentStatement)
	if ifStatement.AlternateStatement != nil {
		printer.Writer.Write(" else ")
		printer.writeStatement(ifStatement.AlternateStatement)
	}
}

func (printer *Printer) writeBlockStatement(blockStatement *ast.BlockStatement) {
	printer.Writer.Write("{\n")
	printer.increaseIndent()
	printer.writeStatementList(blockStatement.Body)
	printer.decreaseIndent()
	printer.Writer.Write("\n}")
}

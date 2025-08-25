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
	for index, statement := range statementList {
		if index > 0 {
			printer.Writer.Write("\n")
		}
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
	case ast.NodeTypeIdentifier:
		printer.writeIdentifier(expression.AsIdentifier())
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
		case ast.NodeTypeMemberExpression:
			printer.writeMemberExpression(callExpression.Callee.AsMemberExpression())
		}
		printer.Writer.Write("(")
		for index, expression := range callExpression.Args {
			printer.writeExpression(expression)
			if index < len(callExpression.Args)-1 {
				printer.Writer.Write(", ")
			}
		}
		printer.Writer.Write(")")
	case ast.NodeTypeBinaryExpression:
		binaryExpression := expression.AsBinaryExpression()
		printer.writeExpression(binaryExpression.Left)
		printer.Writer.Writef(" %s ", binaryExpression.Operator)
		printer.writeExpression(binaryExpression.Right)
	case ast.NodeTypeArrayExpression:
		arrayExpression := expression.AsArrayExpression()
		printer.Writer.Write("[")
		for index, element := range arrayExpression.Elements {
			if element.Type == ast.NodeTypeSpreadElement {
				printer.writeSpreadElement(element.AsSpreadElement())
			} else {
				printer.writeExpression(element)
			}
			if index < len(arrayExpression.Elements)-1 {
				printer.Writer.Write(", ")
			}
		}
		printer.Writer.Write("]")
	case ast.NodeTypeObjectExpression:
		objectExpression := expression.AsObjectExpression()
		printer.Writer.Write("{")
		printer.increaseIndent()
		for index, property := range objectExpression.Properties {
			printer.Writer.Write("\n")
			printer.writeIndent()
			switch property.Type {
			case ast.NodeTypeSpreadElement:
				printer.writeSpreadElement(property.AsSpreadElement())
			case ast.NodeTypeObjectProperty:
				objectProperty := property.AsObjectProperty()
				switch objectProperty.Key.Type {
				case ast.NodeTypeIdentifier:
					printer.writeIdentifier(objectProperty.Key.AsIdentifier())
				case ast.NodeTypeDecimalLiteral:
					printer.Writer.Write(objectProperty.Key.AsDecimalLiteral().Value)
				case ast.NodeTypeStringLiteral:
					printer.Writer.Writef("\"%s\"", objectProperty.Key.AsStringLiteral().Value)
				}
				printer.Writer.Write(": ")
				printer.writeExpression(objectProperty.Value)
			}

			if index < len(objectExpression.Properties)-1 {
				printer.Writer.Write(",")
			}
		}
		printer.decreaseIndent()
		if len(objectExpression.Properties) > 0 {
			printer.Writer.Write("\n")
		}
		printer.writeIndent()
		printer.Writer.Write("}")

	case ast.NodeTypeMemberExpression:
		printer.writeMemberExpression(expression.AsMemberExpression())
	}
}

func (printer *Printer) writeIdentifier(identifier *ast.Identifier) {
	printer.Writer.Write(identifier.Name)
}

func (printer *Printer) writeSpreadElement(spreadElement *ast.SpreadElement) {
	printer.Writer.Write("...")
	printer.writeExpression(spreadElement.Argument)
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
	printer.Writer.Write("\n")
	printer.Writer.Write(strings.Repeat(" ", printer.indent))
	printer.Writer.Write("}")
}

func (printer *Printer) writeIndent() {
	printer.Writer.Write(strings.Repeat(" ", printer.indent))
}

func (printer *Printer) writeMemberExpression(memberExpression *ast.MemberExpression) {
	printer.writeExpression(memberExpression.Object)
	printer.Writer.Write(".")
	printer.writeExpression(memberExpression.Property)
}

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

func WriteSourceFile(sourceFile *ast.SourceFile) string {
	printer := NewPrinter()
	printer.Write(sourceFile)
	return printer.Writer.Output
}

func (printer *Printer) increaseIndent() { printer.indent += 2 }
func (printer *Printer) decreaseIndent() { printer.indent -= 2 }

func (printer *Printer) Write(node *ast.SourceFile) {
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
	printer.writeIndent()
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
	case ast.NodeTypeFunctionDeclaration:
		printer.writeFunctionDeclaration(statement.AsFunctionDeclaration())
	case ast.NodeTypeReturnStatement:
		printer.writeReturnStatement(statement.AsReturnStatement())
		printer.Writer.Write(";")
	case ast.NodeTypeForStatement:
		printer.writeForStatement(statement.AsForStatement())
	case ast.NodeTypeImportDeclaration:
		importDeclaration := statement.AsImportDeclaration()
		printer.Writer.Write("import ")
		didWriteOpenBracker := false
		for index, specifier := range importDeclaration.Specifiers {
			switch specifier.Type {
			case ast.NodeTypeImportDefaultSpecifier:
				printer.Writer.Write(specifier.AsImportDefaultSpecifier().Local.Name)
			case ast.NodeTypeImportSpecifier:
				if !didWriteOpenBracker {
					printer.Writer.Write("{ ")
					didWriteOpenBracker = true
				}
				printer.Writer.Write(specifier.AsImportSpecifier().Local.Name)
			case ast.NodeTypeImportNamespaceSpecifier:
				printer.Writer.Writef("* as %s", specifier.AsImportNamespaceSpecifier().Local.Name)
			}

			if index < len(importDeclaration.Specifiers)-1 {
				printer.Writer.Write(", ")
			}
		}

		if didWriteOpenBracker {
			printer.Writer.Write(" }")
		}

		printer.Writer.Write(" from ")
		printer.Writer.Writef("\"%s\"", importDeclaration.Source.Value)
		printer.Writer.Write(";")
	}
}

func (printer *Printer) writeExpressionStatement(expressionStatement *ast.ExpressionStatement) {
	printer.writeExpression(expressionStatement.Expression)
	printer.Writer.Write(";")
}

func (printer *Printer) writeVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Declare {
		return
	}

	printer.Writer.Write(fmt.Sprintf("let %s = ", variableDeclaration.Identifier.Name))
	if variableDeclaration.Initializer != nil {
		printer.writeInitializer(variableDeclaration.Initializer)
	}
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
	case ast.NodeTypeUpdateExpression:
		updateExpression := expression.AsUpdateExpression()

		if updateExpression.Prefix {
			printer.Writer.Write(updateExpression.Operator)
		}
		printer.writeExpression(updateExpression.Argument)
		if !updateExpression.Prefix {
			printer.Writer.Write(updateExpression.Operator)
		}
	case ast.NodeTypeAssignmentExpression:
		printer.writeAssignmentExpression(expression.AsAssignmentExpression())
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
	// FIXME: Find a better way. writeStatement adds indent.
	// 		  We check for block statements here to prevent extra indent.
	if ifStatement.ConsequentStatement.Type == ast.NodeTypeBlockStatement {
		printer.writeBlockStatement(ifStatement.ConsequentStatement.AsBlockStatement())
	} else {
		printer.writeStatement(ifStatement.ConsequentStatement)
	}
	if ifStatement.AlternateStatement != nil {
		printer.Writer.Write(" else ")
		printer.writeStatement(ifStatement.AlternateStatement)
	}
}

func (printer *Printer) writeBlockStatement(blockStatement *ast.BlockStatement) {
	if len(blockStatement.Body) == 0 {
		printer.Writer.Write("{}")
		return
	}

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
	if memberExpression.Computed {
		printer.Writer.Write("[")
	} else {
		printer.Writer.Write(".")
	}
	printer.writeExpression(memberExpression.Property)
	if memberExpression.Computed {
		printer.Writer.Write("]")
	}
}

func (printer *Printer) writeFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	printer.Writer.Write("function ")
	printer.Writer.Write(functionDeclaration.ID.Name)
	printer.Writer.Write("(")
	for index, param := range functionDeclaration.Params {
		switch param.Type {
		case ast.NodeTypeRestElement:
			printer.Writer.Write("...")
			printer.Writer.Write(param.AsRestElement().Argument.Name)
		case ast.NodeTypeIdentifier:
			printer.Writer.Write(param.AsIdentifier().Name)
		}

		if index < len(functionDeclaration.Params)-1 {
			printer.Writer.Write(", ")
		}
	}
	printer.Writer.Write(") ")
	printer.writeBlockStatement(functionDeclaration.Body)
}

func (printer *Printer) writeReturnStatement(returnStatement *ast.ReturnStatement) {
	printer.Writer.Write("return ")
	printer.writeExpression(returnStatement.Argument)
}

func (printer *Printer) writeForStatement(forStatement *ast.ForStatement) {
	printer.Writer.Write("for (")
	if forStatement.Init.Type == ast.NodeTypeVariableDeclaration {
		printer.writeVariableDeclaration(forStatement.Init.AsVariableDeclaration())
	} else {
		printer.writeExpression(forStatement.Init)
	}
	printer.Writer.Write("; ")
	printer.writeExpression(forStatement.Test)
	printer.Writer.Write("; ")
	printer.writeExpression(forStatement.Update)
	printer.Writer.Write(") ")
	printer.writeBlockStatement(forStatement.Body.AsBlockStatement())
}

func (printer *Printer) writeAssignmentExpression(assignmentExpression *ast.AssignmentExpression) {
	printer.writeExpression(assignmentExpression.Left)
	printer.Writer.Writef(" %s ", assignmentExpression.Operator)
	printer.writeExpression(assignmentExpression.Right)
}

package emitter

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

type Emitter struct {
	Writer *Writer
	indent int
}

func NewEmitter() *Emitter { return &Emitter{Writer: NewWriter(), indent: 0} }

func EmitSourceFile(sourceFile *ast.SourceFile) string {
	emitter := NewEmitter()
	emitter.Emit(sourceFile)
	return emitter.Writer.Output
}

func (emitter *Emitter) increaseIndent() { emitter.indent += 2 }
func (emitter *Emitter) decreaseIndent() { emitter.indent -= 2 }

func (emitter *Emitter) Emit(node *ast.SourceFile) {
	emitter.emitStatementList(node.Body)
}

func (emitter *Emitter) emitStatementList(statementList []*ast.Node) {
	for index, statement := range statementList {
		if index > 0 {
			emitter.Writer.Write("\n")
		}
		emitter.emitStatement(statement)
	}
}

func (emitter *Emitter) emitStatement(statement *ast.Node) {
	emitter.emitIndent()
	switch statement.Type {
	case ast.NodeTypeVariableDeclaration:
		emitter.emitVariableDeclaration(statement.AsVariableDeclaration())
		emitter.Writer.Write(";")
	case ast.NodeTypeIfStatement:
		emitter.emitIfStatement(statement.AsIfStatement())
	case ast.NodeTypeBlockStatement:
		emitter.emitBlockStatement(statement.AsBlockStatement())
	case ast.NodeTypeExpressionStatement:
		emitter.emitExpressionStatement(statement.AsExpressionStatement())
	case ast.NodeTypeFunctionDeclaration:
		emitter.emitFunctionDeclaration(statement.AsFunctionDeclaration())
	case ast.NodeTypeReturnStatement:
		emitter.emitReturnStatement(statement.AsReturnStatement())
		emitter.Writer.Write(";")
	case ast.NodeTypeForStatement:
		emitter.emitForStatement(statement.AsForStatement())
	case ast.NodeTypeImportDeclaration:
		importDeclaration := statement.AsImportDeclaration()
		emitter.Writer.Write("import ")
		didEmitOpenBracker := false
		for index, specifier := range importDeclaration.Specifiers {
			switch specifier.Type {
			case ast.NodeTypeImportDefaultSpecifier:
				emitter.Writer.Write(specifier.AsImportDefaultSpecifier().Local.Name)
			case ast.NodeTypeImportSpecifier:
				if !didEmitOpenBracker {
					emitter.Writer.Write("{ ")
					didEmitOpenBracker = true
				}
				emitter.Writer.Write(specifier.AsImportSpecifier().Local.Name)
			case ast.NodeTypeImportNamespaceSpecifier:
				emitter.Writer.Writef("* as %s", specifier.AsImportNamespaceSpecifier().Local.Name)
			}

			if index < len(importDeclaration.Specifiers)-1 {
				emitter.Writer.Write(", ")
			}
		}

		if didEmitOpenBracker {
			emitter.Writer.Write(" }")
		}

		emitter.Writer.Write(" from ")
		emitter.Writer.Writef("\"%s\"", importDeclaration.Source.Value)
		emitter.Writer.Write(";")
	}
}

func (emitter *Emitter) emitExpressionStatement(expressionStatement *ast.ExpressionStatement) {
	emitter.emitExpression(expressionStatement.Expression)
	emitter.Writer.Write(";")
}

func (emitter *Emitter) emitVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Declare {
		return
	}

	emitter.Writer.Write(fmt.Sprintf("let %s = ", variableDeclaration.Identifier.Name))
	if variableDeclaration.Initializer != nil {
		emitter.emitInitializer(variableDeclaration.Initializer)
	}
}

func (emitter *Emitter) emitInitializer(initializer *ast.Initializer) {
	emitter.emitExpression(initializer.Expression)
}

func (emitter *Emitter) emitExpression(expression *ast.Node) {
	switch expression.Type {
	case ast.NodeTypeIdentifier:
		emitter.emitIdentifier(expression.AsIdentifier())
	case ast.NodeTypeDecimalLiteral:
		emitter.Writer.Write(expression.AsDecimalLiteral().Value)
	case ast.NodeTypeStringLiteral:
		emitter.Writer.Writef("\"%s\"", expression.AsStringLiteral().Value)
	case ast.NodeTypeNullLiteral:
		emitter.Writer.Write("null")
	case ast.NodeTypeBooleanLiteral:
		if expression.AsBooleanLiteral().Value {
			emitter.Writer.Write("true")
		} else {
			emitter.Writer.Write("false")
		}
	case ast.NodeTypeCallExpression:
		callExpression := expression.AsCallExpression()
		switch callExpression.Callee.Type {
		case ast.NodeTypeIdentifier:
			emitter.Writer.Write(callExpression.Callee.AsIdentifier().Name)
		case ast.NodeTypeMemberExpression:
			emitter.emitMemberExpression(callExpression.Callee.AsMemberExpression())
		}
		emitter.Writer.Write("(")
		for index, expression := range callExpression.Args {
			emitter.emitExpression(expression)
			if index < len(callExpression.Args)-1 {
				emitter.Writer.Write(", ")
			}
		}
		emitter.Writer.Write(")")
	case ast.NodeTypeBinaryExpression:
		binaryExpression := expression.AsBinaryExpression()
		emitter.emitExpression(binaryExpression.Left)
		emitter.Writer.Writef(" %s ", binaryExpression.Operator)
		emitter.emitExpression(binaryExpression.Right)
	case ast.NodeTypeArrayExpression:
		arrayExpression := expression.AsArrayExpression()
		emitter.Writer.Write("[")
		for index, element := range arrayExpression.Elements {
			if element.Type == ast.NodeTypeSpreadElement {
				emitter.emitSpreadElement(element.AsSpreadElement())
			} else {
				emitter.emitExpression(element)
			}
			if index < len(arrayExpression.Elements)-1 {
				emitter.Writer.Write(", ")
			}
		}
		emitter.Writer.Write("]")
	case ast.NodeTypeObjectExpression:
		objectExpression := expression.AsObjectExpression()
		emitter.Writer.Write("{")
		emitter.increaseIndent()
		for index, property := range objectExpression.Properties {
			emitter.Writer.Write("\n")
			emitter.emitIndent()
			switch property.Type {
			case ast.NodeTypeSpreadElement:
				emitter.emitSpreadElement(property.AsSpreadElement())
			case ast.NodeTypeObjectProperty:
				objectProperty := property.AsObjectProperty()
				switch objectProperty.Key.Type {
				case ast.NodeTypeIdentifier:
					emitter.emitIdentifier(objectProperty.Key.AsIdentifier())
				case ast.NodeTypeDecimalLiteral:
					emitter.Writer.Write(objectProperty.Key.AsDecimalLiteral().Value)
				case ast.NodeTypeStringLiteral:
					emitter.Writer.Writef("\"%s\"", objectProperty.Key.AsStringLiteral().Value)
				}
				emitter.Writer.Write(": ")
				emitter.emitExpression(objectProperty.Value)
			}

			if index < len(objectExpression.Properties)-1 {
				emitter.Writer.Write(",")
			}
		}
		emitter.decreaseIndent()
		if len(objectExpression.Properties) > 0 {
			emitter.Writer.Write("\n")
		}
		emitter.emitIndent()
		emitter.Writer.Write("}")

	case ast.NodeTypeMemberExpression:
		emitter.emitMemberExpression(expression.AsMemberExpression())
	case ast.NodeTypeUpdateExpression:
		updateExpression := expression.AsUpdateExpression()

		if updateExpression.Prefix {
			emitter.Writer.Write(updateExpression.Operator)
		}
		emitter.emitExpression(updateExpression.Argument)
		if !updateExpression.Prefix {
			emitter.Writer.Write(updateExpression.Operator)
		}
	case ast.NodeTypeAssignmentExpression:
		emitter.emitAssignmentExpression(expression.AsAssignmentExpression())
	case ast.NodeTypeThisEpxression:
		emitter.Writer.Write("this")
	case ast.NodeTypeFunctionExpression:
		emitter.emitFunctionExpression(expression.AsFunctionExpression())
	}
}

func (emitter *Emitter) emitIdentifier(identifier *ast.Identifier) {
	emitter.Writer.Write(identifier.Name)
}

func (emitter *Emitter) emitSpreadElement(spreadElement *ast.SpreadElement) {
	emitter.Writer.Write("...")
	emitter.emitExpression(spreadElement.Argument)
}

func (emitter *Emitter) emitIfStatement(ifStatement *ast.IfStatement) {
	emitter.Writer.Write("if (")
	emitter.emitExpression(ifStatement.TestExpression)
	emitter.Writer.Write(") ")
	// FIXME: Find a better way. emitStatement adds indent.
	// 		  We check for block statements here to prevent extra indent.
	if ifStatement.ConsequentStatement.Type == ast.NodeTypeBlockStatement {
		emitter.emitBlockStatement(ifStatement.ConsequentStatement.AsBlockStatement())
	} else {
		emitter.emitStatement(ifStatement.ConsequentStatement)
	}
	if ifStatement.AlternateStatement != nil {
		emitter.Writer.Write(" else ")
		emitter.emitStatement(ifStatement.AlternateStatement)
	}
}

func (emitter *Emitter) emitBlockStatement(blockStatement *ast.BlockStatement) {
	if len(blockStatement.Body) == 0 {
		emitter.Writer.Write("{}")
		return
	}

	emitter.Writer.Write("{\n")
	emitter.increaseIndent()
	emitter.emitStatementList(blockStatement.Body)
	emitter.decreaseIndent()
	emitter.Writer.Write("\n")
	emitter.Writer.Write(strings.Repeat(" ", emitter.indent))
	emitter.Writer.Write("}")
}

func (emitter *Emitter) emitIndent() {
	emitter.Writer.Write(strings.Repeat(" ", emitter.indent))
}

func (emitter *Emitter) emitMemberExpression(memberExpression *ast.MemberExpression) {
	emitter.emitExpression(memberExpression.Object)
	if memberExpression.Computed {
		emitter.Writer.Write("[")
	} else {
		emitter.Writer.Write(".")
	}
	emitter.emitExpression(memberExpression.Property)
	if memberExpression.Computed {
		emitter.Writer.Write("]")
	}
}

func (emitter *Emitter) emitFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	emitter.Writer.Write("function ")
	emitter.Writer.Write(functionDeclaration.ID.Name)
	emitter.Writer.Write("(")
	for index, param := range functionDeclaration.Params {
		if param.Type == ast.NodeTypeBindingElement {
			bindingElement := param.AsBindingElement()
			if bindingElement.Rest {
				emitter.Writer.Write("...")
			}

			if bindingElement.Element != nil {
				switch bindingElement.Element.Type {
				case ast.NodeTypeIdentifier:
					emitter.Writer.Write(bindingElement.Element.AsIdentifier().Name)
				}
			}
		}

		if index < len(functionDeclaration.Params)-1 {
			emitter.Writer.Write(", ")
		}
	}
	emitter.Writer.Write(") ")
	emitter.emitBlockStatement(functionDeclaration.Body)
}

func (emitter *Emitter) emitReturnStatement(returnStatement *ast.ReturnStatement) {
	emitter.Writer.Write("return ")
	emitter.emitExpression(returnStatement.Argument)
}

func (emitter *Emitter) emitForStatement(forStatement *ast.ForStatement) {
	emitter.Writer.Write("for (")
	if forStatement.Init.Type == ast.NodeTypeVariableDeclaration {
		emitter.emitVariableDeclaration(forStatement.Init.AsVariableDeclaration())
	} else {
		emitter.emitExpression(forStatement.Init)
	}
	emitter.Writer.Write("; ")
	emitter.emitExpression(forStatement.Test)
	emitter.Writer.Write("; ")
	emitter.emitExpression(forStatement.Update)
	emitter.Writer.Write(") ")
	emitter.emitBlockStatement(forStatement.Body.AsBlockStatement())
}

func (emitter *Emitter) emitAssignmentExpression(assignmentExpression *ast.AssignmentExpression) {
	emitter.emitExpression(assignmentExpression.Left)
	emitter.Writer.Writef(" %s ", assignmentExpression.Operator)
	emitter.emitExpression(assignmentExpression.Right)
}

func (emitter *Emitter) emitFunctionExpression(functionExpression *ast.FunctionExpression) {
	emitter.Writer.Write("function ")
	if functionExpression.ID != nil {
		emitter.Writer.Write(functionExpression.ID.Name)
	}
	emitter.Writer.Write("(")
	for index, param := range functionExpression.Params {
		switch param.Type {
		case ast.NodeTypeRestElement:
			emitter.Writer.Write("...")
			emitter.Writer.Write(param.AsRestElement().Argument.Name)
		case ast.NodeTypeIdentifier:
			emitter.Writer.Write(param.AsIdentifier().Name)
		}

		if index < len(functionExpression.Params)-1 {
			emitter.Writer.Write(", ")
		}
	}
	emitter.Writer.Write(") ")
	emitter.emitBlockStatement(functionExpression.Body)
}

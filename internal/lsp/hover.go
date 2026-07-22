package lsp

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/emitter"
	"fmt"
	"strings"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func ProvideHover(node *ast.Node, nameResolver *binder.NameResolver) (result *defines.Hover, err error) {
	// If it's not an identifier, skip
	if node.Type != ast.NodeTypeIdentifier {
		return nil, nil
	}

	contents := ""

	switch node.Parent.Type {
	case ast.NodeTypeVariableDeclaration:
		variableDeclaration := node.Parent.AsVariableDeclaration()
		if variableDeclaration.Name != node {
			break
		}
		contents = formatVarHover(variableDeclaration)

	case ast.NodeTypeFunctionDeclaration:
		functionDeclaration := node.Parent.AsFunctionDeclaration()
		if functionDeclaration.ID != node.AsIdentifier() {
			break
		}
		contents = formatFuncHover(functionDeclaration)

	case ast.NodeTypeMemberExpression:
		memberExpression := node.Parent.AsMemberExpression()
		if memberExpression.Object == node {
			symbol := nameResolver.Resolve(node.AsIdentifier().Name, node)
			if symbol == nil {
				return nil, nil
			}
			contents = formatDeclarationHover(symbol.Node)
		}
		if memberExpression.Property != node {
			break
		}
		// FIXME: Handle property

	default:
		symbol := nameResolver.Resolve(node.AsIdentifier().Name, node)
		if symbol == nil {
			return nil, nil
		}
		contents = formatDeclarationHover(symbol.Node)
	}

	if contents == "" {
		return nil, nil
	}

	return &defines.Hover{
		Contents: defines.MarkupContent{
			Kind:  "markdown",
			Value: contents,
		},
	}, nil
}

// Formats hover text for a node resolved from a symbol (variable or function)
func formatDeclarationHover(node *ast.Node) string {
	switch node.Type {
	case ast.NodeTypeVariableDeclaration:
		return formatVarHover(node.AsVariableDeclaration())
	case ast.NodeTypeFunctionDeclaration:
		return formatFuncHover(node.AsFunctionDeclaration())
	default:
		return ""
	}
}

// Formats a variable declaration for hover display
func formatVarHover(decl *ast.VariableDeclaration) string {
	identifier := decl.Name.AsIdentifier()

	var prefix string
	if decl.Symbol != nil && decl.Symbol.OriginalName != nil {
		prefix = fmt.Sprintf("// الإسم الأصلي: %s\n", *decl.Symbol.OriginalName)
	}

	return fmt.Sprintf("```arts\n%sمتغير %s: %s\n```", prefix, identifier.Name, emitter.EmitTypeNode(decl.TypeNode()))
}

// Formats a function declaration for hover display
func formatFuncHover(decl *ast.FunctionDeclaration) string {
	identifier := decl.ID

	var prefix string
	if decl.Symbol != nil && decl.Symbol.OriginalName != nil {
		prefix = fmt.Sprintf("// الإسم الأصلي: %s\n", *decl.Symbol.OriginalName)
	}

	var params []string
	for _, param := range decl.Params {
		parameter := param.AsParameter()
		params = append(params, fmt.Sprintf("%s: %s", parameter.Name.AsIdentifier().Name, emitter.EmitTypeNode(parameter.TypeNode())))
	}

	return fmt.Sprintf("```arts\n%sدالة %s(%s)\n```", prefix, identifier.Name, strings.Join(params, ", "))
}

package lsp

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
	"fmt"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func ProvideHover(node *ast.Node, nameResolver *binder.NameResolver) (result *defines.Hover, err error) {
	// If it identifier
	if node.Type != ast.NodeTypeIdentifier {
		return nil, nil
	}

	contents := ""

	switch node.Parent.Type {
	case ast.NodeTypeVariableDeclaration:
		// If it variable name
		variableDeclaration := node.Parent.AsVariableDeclaration()
		if variableDeclaration.Name != node {
			break
		}
		contents = fmt.Sprintf("```arts\nمتغير %s: %s\n```", node.AsIdentifier().Name, TypeNodeToString(node.TypeNode()))
	case ast.NodeTypeFunctionDeclaration:
		// If it function name
		functionDeclaration := node.Parent.AsFunctionDeclaration()
		if functionDeclaration.ID != node.AsIdentifier() {
			break
		}

		contents = fmt.Sprintf("```arts\nدالة %s(", node.AsIdentifier().Name)
		for index, param := range functionDeclaration.Params {
			parameter := param.AsParameter()
			paramStr := fmt.Sprintf("%s: %s", parameter.Name.AsIdentifier().Name, TypeNodeToString(parameter.TypeNode()))
			if len(functionDeclaration.Params)-1 != index {
				paramStr = fmt.Sprintf("%s,", paramStr)
			}
			contents = fmt.Sprintf("%s%s", contents, paramStr)
		}
		contents = fmt.Sprintf("%s)\n```", contents)
	case ast.NodeTypeMemberExpression:
		// If it property name
		memberExpression := node.Parent.AsMemberExpression()
		if memberExpression.Property != node {
			break
		}
		// FIXME: Handle property

	default:
		// If it refrencing identifier
		symbol := nameResolver.Resolve(node.AsIdentifier().Name, node)
		if symbol == nil {
			return nil, nil
		}

		switch symbol.Node.Type {
		case ast.NodeTypeVariableDeclaration:
			// If it variable name
			variableDeclaration := symbol.Node.AsVariableDeclaration()
			identifier := variableDeclaration.Name.AsIdentifier()
			contents = fmt.Sprintf("```arts\nمتغير %s: %s\n```", identifier.Name, TypeNodeToString(identifier.TypeNode()))
		case ast.NodeTypeFunctionDeclaration:
			// If it function name
			functionDeclaration := symbol.Node.AsFunctionDeclaration()
			identifier := functionDeclaration.ID

			contents = fmt.Sprintf("```arts\nدالة %s(", identifier.Name)
			for index, param := range functionDeclaration.Params {
				parameter := param.AsParameter()
				paramStr := fmt.Sprintf("%s: %s", parameter.Name.AsIdentifier().Name, TypeNodeToString(parameter.TypeNode()))
				if len(functionDeclaration.Params)-1 != index {
					paramStr = fmt.Sprintf("%s,", paramStr)
				}
				contents = fmt.Sprintf("%s%s", contents, paramStr)
			}
			contents = fmt.Sprintf("%s)\n```", contents)
		default:
			return nil, nil
		}
	}

	return &defines.Hover{
		Contents: defines.MarkupContent{
			Kind:  "markdown",
			Value: contents,
		},
	}, nil
}

func TypeNodeToString(node *ast.Node) string {
	if node == nil {
		return "أي_نوع"
	}

	switch node.Type {
	case ast.NodeTypeTypeAnnotation:
		return TypeNodeToString(node.AsTypeAnnotation().TypeAnnotation)
	case ast.NodeTypeStringKeyword:
		return "نص"
	case ast.NodeTypeNumberKeyword:
		return "عدد"
	case ast.NodeTypeBooleanKeyword:
		return "قيمة_منطقية"
	case ast.NodeTypeNullKeyword:
		return "فارغ"
	case ast.NodeTypeAnyKeyword:
		return "أي_نوع"
	case ast.NodeTypeTypeReference:
		return node.AsTypeReferenceNode().TypeName.Name
	case ast.NodeTypeArrayType:
		return fmt.Sprintf("%s[]", TypeNodeToString(node.AsArrayType().ElementType))
	case ast.NodeTypeUnionType:
		unionType := node.AsUnionType()
		unionTypeString := ""
		for index, _type := range unionType.Types {
			unionTypeString = fmt.Sprintf("%s%s", unionTypeString, TypeNodeToString(_type))
			if index != len(unionType.Types)-1 {
				unionTypeString = fmt.Sprintf("%s | ", unionTypeString)
			}
		}
		return unionTypeString
	default:
		return "أي_نوع"
	}
}

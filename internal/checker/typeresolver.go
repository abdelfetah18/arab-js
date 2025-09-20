package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
)

type TypeResolver struct {
	NameResolver *binder.NameResolver
}

func NewTypeResolver(nameResolver *binder.NameResolver) *TypeResolver {
	return &TypeResolver{
		NameResolver: nameResolver,
	}
}

func (t *TypeResolver) ResolveTypeFromNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeIdentifier:
		identifier := node.AsIdentifier()
		if node.Parent != nil && node.Parent.Type == ast.NodeTypeVariableDeclaration {
			return t.ResolveTypeAnnotation(identifier.TypeAnnotation)
		}

		if node.Parent != nil && node.Parent.Type == ast.NodeTypeFunctionDeclaration {
			return t.ResolveTypeAnnotation(identifier.TypeAnnotation)
		}

		symbol := t.NameResolver.Resolve(identifier.Name, node)
		if symbol == nil {
			return nil
		}

		return t.ResolveTypeFromNode(symbol.Node)
	case ast.NodeTypeVariableDeclaration:
		return t.ResolveTypeAnnotation(node.AsVariableDeclaration().Identifier.TypeAnnotation)
	case ast.NodeTypeFunctionDeclaration:
		functionDeclaration := node.AsFunctionDeclaration()
		functionType := NewType(NewFunctionType())
		for _, param := range functionDeclaration.Params {
			functionType.AddParamType(t.ResolveTypeFromNode(param))
		}
		functionType.ReturnType = t.ResolveTypeAnnotation(functionDeclaration.TypeAnnotation)
		return functionType.AsType()
	case ast.NodeTypeStringLiteral:
		return NewType(NewStringType()).AsType()
	case ast.NodeTypeDecimalLiteral:
		return NewType(NewNumberType()).AsType()
	case ast.NodeTypeBooleanLiteral:
		return NewType(NewBooleanType()).AsType()
	case ast.NodeTypeNullLiteral:
		return NewType(NewNullType()).AsType()
	default:
		return nil
	}
}

func (t *TypeResolver) ResolveTypeAnnotation(typeAnnotation *ast.TypeAnnotation) *Type {
	if typeAnnotation == nil {
		return nil
	}

	return t.ResolveTypeNode(typeAnnotation.TypeAnnotation)
}

func (t *TypeResolver) ResolveTypeNode(typeNode *ast.Node) *Type {
	if typeNode == nil {
		return nil
	}

	switch typeNode.Type {
	case ast.NodeTypeStringKeyword:
		return NewType(NewStringType()).AsType()
	case ast.NodeTypeBooleanKeyword:
		return NewType(NewBooleanType()).AsType()
	case ast.NodeTypeNullKeyword:
		return NewType(NewNullType()).AsType()
	case ast.NodeTypeNumberKeyword:
		return NewType(NewNumberType()).AsType()
	case ast.NodeTypeTypeLiteral:
		objectType := NewObjectType()
		typeLiteral := typeNode.AsTypeLiteral()
		for _, member := range typeLiteral.Members {
			propertySignature := member.AsPropertySignature()
			switch propertySignature.Key.Type {
			case ast.NodeTypeIdentifier:
				identifier := propertySignature.Key.AsIdentifier()
				objectType.AddProperty(
					identifier.Name,
					&PropertyType{
						Type:         t.ResolveTypeNode(propertySignature.TypeNode()),
						Name:         identifier.Name,
						OriginalName: identifier.OriginalName,
					},
				)
			}
		}

		return NewType(objectType).AsType()
	case ast.NodeTypeTypeReference:
		return t.ResolveTypeFromTypeReference(typeNode.AsTypeReference())
	case ast.NodeTypeFunctionType:
		functionTypeNode := typeNode.AsFunctionType()
		functionType := NewFunctionType()
		for _, param := range functionTypeNode.Params {
			functionType.AddParamType(t.ResolveTypeFromNode(param))
		}
		functionType.ReturnType = t.ResolveTypeAnnotation(functionTypeNode.TypeAnnotation)
		return NewType(functionType).AsType()
	}

	return nil
}

func (t *TypeResolver) ResolveTypeFromTypeReference(typeReference *ast.TypeReference) *Type {
	symbol := t.NameResolver.Resolve(typeReference.TypeName.Name, typeReference.TypeName.AsNode())
	if symbol == nil {
		return nil
	}

	return t.ResolveTypeFromTypeDeclaration(symbol.Node)
}

func (t *TypeResolver) ResolveTypeFromTypeDeclaration(typeDeclaration *ast.Node) *Type {
	switch typeDeclaration.Type {
	case ast.NodeTypeInterfaceDeclaration:
		interfaceDeclaration := typeDeclaration.AsInterfaceDeclaration()
		objectType := NewObjectType()
		for _, member := range interfaceDeclaration.Body.Body {
			propertySignature := member.AsPropertySignature()
			switch propertySignature.Key.Type {
			case ast.NodeTypeIdentifier:
				identifier := propertySignature.Key.AsIdentifier()
				objectType.AddProperty(
					identifier.Name,
					&PropertyType{
						Type:         t.ResolveTypeNode(propertySignature.TypeNode()),
						Name:         identifier.Name,
						OriginalName: identifier.OriginalName,
					},
				)
			}
		}

		return NewType(objectType).AsType()
	case ast.NodeTypeTypeAliasDeclaration:
		typeAliasDeclaration := typeDeclaration.AsTypeAliasDeclaration()
		return t.ResolveTypeAnnotation(typeAliasDeclaration.TypeAnnotation)
	default:
		return nil
	}
}

package ast

import "encoding/json"

type NodeType int16

const (
	NodeTypeUnknown NodeType = iota
	NodeTypeExpressionStatement
	NodeTypeVariableDeclaration
	NodeTypeIdentifier
	NodeTypeInitializer
	NodeTypeStringLiteral
	NodeTypeNullLiteral
	NodeTypeBooleanLiteral
	NodeTypeDecimalLiteral
	NodeTypeIfStatement
	NodeTypeBlockStatement
	NodeTypeReturnStatement
	NodeTypeForStatement
	NodeTypeUpdateExpression
	NodeTypeAssignmentExpression
	NodeTypeFunctionDeclaration
	NodeTypeCallExpression
	NodeTypeThisEpxression
	NodeTypeFunctionExpression
	NodeTypeMemberExpression
	NodeTypeImportSpecifier
	NodeTypeImportDefaultSpecifier
	NodeTypeImportNamespaceSpecifier
	NodeTypeImportDeclaration
	NodeTypeExportNamedDeclaration
	NodeTypeExportDefaultDeclaration
	NodeTypeExportSpecifier
	NodeTypeBinaryExpression
	NodeTypeSpreadElement
	NodeTypeArrayExpression
	NodeTypeObjectProperty
	NodeTypeObjectMethod
	NodeTypeObjectExpression
	NodeTypeDirectiveLiteral
	NodeTypeDirective
	NodeTypeSourceFile

	NodeTypeTypeAnnotation
	NodeTypeStringKeyword
	NodeTypeNumberKeyword
	NodeTypeBooleanKeyword
	NodeTypeNullKeyword
	NodeTypeAnyKeyword
	NodeTypeInterfaceDeclaration
	NodeTypeInterfaceBody
	NodeTypePropertySignature
	NodeTypeFunctionType
	NodeTypeTypeLiteral
	NodeTypeTypeAliasDeclaration
	NodeTypeTypeReference
	NodeTypeArrayType
	NodeTypeRestElement
	NodeTypeModuleDeclaration
	NodeTypeModuleBlock
	NodeTypeUnionType
	NodeTypeTypeParameter
	NodeTypeTypeParametersDeclaration
	NodeTypeTypeParameterInstantiation
	NodeTypeIndexSignatureDeclaration
)

func (t NodeType) String() string {
	switch t {
	case NodeTypeUnknown:
		return "Unknown"
	case NodeTypeExpressionStatement:
		return "ExpressionStatement"
	case NodeTypeVariableDeclaration:
		return "VariableDeclaration"
	case NodeTypeIdentifier:
		return "Identifier"
	case NodeTypeInitializer:
		return "Initializer"
	case NodeTypeStringLiteral:
		return "StringLiteral"
	case NodeTypeNullLiteral:
		return "NullLiteral"
	case NodeTypeBooleanLiteral:
		return "BooleanLiteral"
	case NodeTypeDecimalLiteral:
		return "DecimalLiteral"
	case NodeTypeIfStatement:
		return "IfStatement"
	case NodeTypeBlockStatement:
		return "BlockStatement"
	case NodeTypeReturnStatement:
		return "ReturnStatement"
	case NodeTypeForStatement:
		return "ForStatement"
	case NodeTypeUpdateExpression:
		return "UpdateExpression"
	case NodeTypeAssignmentExpression:
		return "AssignmentExpression"
	case NodeTypeFunctionDeclaration:
		return "FunctionDeclaration"
	case NodeTypeCallExpression:
		return "CallExpression"
	case NodeTypeMemberExpression:
		return "MemberExpression"
	case NodeTypeImportSpecifier:
		return "ImportSpecifier"
	case NodeTypeImportDefaultSpecifier:
		return "ImportDefaultSpecifier"
	case NodeTypeImportNamespaceSpecifier:
		return "ImportNamespaceSpecifier"
	case NodeTypeImportDeclaration:
		return "ImportDeclaration"
	case NodeTypeExportNamedDeclaration:
		return "ExportNamedDeclaration"
	case NodeTypeExportDefaultDeclaration:
		return "ExportDefaultDeclaration"
	case NodeTypeExportSpecifier:
		return "ExportSpecifier"
	case NodeTypeBinaryExpression:
		return "BinaryExpression"
	case NodeTypeSpreadElement:
		return "SpreadElement"
	case NodeTypeArrayExpression:
		return "ArrayExpression"
	case NodeTypeObjectProperty:
		return "ObjectProperty"
	case NodeTypeObjectMethod:
		return "ObjectMethod"
	case NodeTypeObjectExpression:
		return "ObjectExpression"
	case NodeTypeDirectiveLiteral:
		return "DirectiveLiteral"
	case NodeTypeDirective:
		return "Directive"
	case NodeTypeSourceFile:
		return "SourceFile"
	case NodeTypeTypeAnnotation:
		return "TypeAnnotation"
	case NodeTypeStringKeyword:
		return "StringKeyword"
	case NodeTypeNumberKeyword:
		return "NumberKeyword"
	case NodeTypeBooleanKeyword:
		return "BooleanKeyword"
	case NodeTypeNullKeyword:
		return "NullKeyword"
	case NodeTypeAnyKeyword:
		return "AnyKeyword"
	case NodeTypeInterfaceDeclaration:
		return "InterfaceDeclaration"
	case NodeTypeInterfaceBody:
		return "InterfaceBody"
	case NodeTypePropertySignature:
		return "PropertySignature"
	case NodeTypeFunctionType:
		return "FunctionType"
	case NodeTypeTypeAliasDeclaration:
		return "TypeAliasDeclaration"
	case NodeTypeTypeReference:
		return "TypeReference"
	case NodeTypeArrayType:
		return "ArrayType"
	case NodeTypeRestElement:
		return "RestElement"
	case NodeTypeTypeLiteral:
		return "TypeLiteral"
	case NodeTypeModuleDeclaration:
		return "ModuleDeclaration"
	case NodeTypeModuleBlock:
		return "ModuleBlock"
	case NodeTypeUnionType:
		return "UnionType"
	case NodeTypeTypeParametersDeclaration:
		return "TypeParametersDeclaration"
	case NodeTypeTypeParameter:
		return "TypeParameter"
	case NodeTypeTypeParameterInstantiation:
		return "TypeParameterInstantiation"
	case NodeTypeThisEpxression:
		return "ThisExpression"
	case NodeTypeFunctionExpression:
		return "FunctionExpression"
	case NodeTypeIndexSignatureDeclaration:
		return "IndexSignatureDeclaration"
	default:
		return "Unknown"
	}
}

func (t NodeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

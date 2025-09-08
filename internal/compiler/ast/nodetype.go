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

	NodeTypeTTypeAnnotation
	NodeTypeTStringKeyword
	NodeTypeTNumberKeyword
	NodeTypeTBooleanKeyword
	NodeTypeTNullKeyword
	NodeTypeTAnyKeyword
	NodeTypeTInterfaceDeclaration
	NodeTypeTInterfaceBody
	NodeTypeTPropertySignature
	NodeTypeTFunctionType
	NodeTypeTTypeAliasDeclaration
	NodeTypeTTypeReference
	NodeTypeTArrayType
	NodeTypeRestElement
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
	case NodeTypeTTypeAnnotation:
		return "TTypeAnnotation"
	case NodeTypeTStringKeyword:
		return "TStringKeyword"
	case NodeTypeTNumberKeyword:
		return "TNumberKeyword"
	case NodeTypeTBooleanKeyword:
		return "TBooleanKeyword"
	case NodeTypeTNullKeyword:
		return "TNullKeyword"
	case NodeTypeTAnyKeyword:
		return "TAnyKeyword"
	case NodeTypeTInterfaceDeclaration:
		return "TInterfaceDeclaration"
	case NodeTypeTInterfaceBody:
		return "TInterfaceBody"
	case NodeTypeTPropertySignature:
		return "TPropertySignature"
	case NodeTypeTFunctionType:
		return "TFunctionType"
	case NodeTypeTTypeAliasDeclaration:
		return "TTypeAliasDeclaration"
	case NodeTypeTTypeReference:
		return "TTypeReference"
	case NodeTypeTArrayType:
		return "TArrayType"
	case NodeTypeRestElement:
		return "RestElement"
	default:
		return "Unknown"
	}
}

func (t NodeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

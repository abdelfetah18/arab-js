package ast

import (
	"encoding/json"
)

type BinaryExpressionOperator = string

const (
	PLUS               BinaryExpressionOperator = "+"
	MINUS              BinaryExpressionOperator = "-"
	SLASH              BinaryExpressionOperator = "/"
	PERCENT            BinaryExpressionOperator = "%"
	STAR               BinaryExpressionOperator = "*"
	COMMA              BinaryExpressionOperator = ","
	DOUBLE_STAR        BinaryExpressionOperator = "**"
	BITWISE_AND        BinaryExpressionOperator = "&"
	BITWISE_OR         BinaryExpressionOperator = "|"
	DOUBLE_RIGHT_ARROW BinaryExpressionOperator = ">>"
	TRIPLE_RIGHT_ARROW BinaryExpressionOperator = ">>>"
	DOUBLE_LEFT_ARROW  BinaryExpressionOperator = "<<"
	BITWISE_XOR        BinaryExpressionOperator = "^"
	EQUAL_EQUAL        BinaryExpressionOperator = "=="
	EQUAL_EQUAL_EQUAL  BinaryExpressionOperator = "==="
	NOT_EQUAL          BinaryExpressionOperator = "!="
	NOT_EQUAL_EQUAL    BinaryExpressionOperator = "!=="
	RIGHT_ARROW        BinaryExpressionOperator = ">"
	LEFT_ARROW         BinaryExpressionOperator = "<"
	RIGHT_ARROW_EQUAL  BinaryExpressionOperator = ">="
	LEFT_ARROW_EQUAL   BinaryExpressionOperator = "<="
)

type UpdateExpressionOperator = string

const (
	PLUS_PLUS   UpdateExpressionOperator = "++"
	MINUS_MINUS UpdateExpressionOperator = "--"
)

type AssignmentExpressionOperator = string

const (
	EQUAL                    AssignmentExpressionOperator = "="
	STAR_EQUAL               AssignmentExpressionOperator = "*="
	SLASH_EQUAL              AssignmentExpressionOperator = "/="
	PERCENT_EQUAL            AssignmentExpressionOperator = "%="
	PLUS_EQUAL               AssignmentExpressionOperator = "+="
	MINUS_EQUAL              AssignmentExpressionOperator = "-="
	DOUBLE_LEFT_ARROW_EQUAL  AssignmentExpressionOperator = "<<="
	DOUBLE_RIGHT_ARROW_EQUAL AssignmentExpressionOperator = ">>="
	AND_EQUAL                AssignmentExpressionOperator = "&="
	XOR_EQUAL                AssignmentExpressionOperator = "^="
	OR_EQUAL                 AssignmentExpressionOperator = "|="
	DOUBLE_STAR_EQUAL        AssignmentExpressionOperator = "**="
)

type Location struct {
	Pos uint `json:"pos"`
	End uint `json:"end"`
}

type Node struct {
	Type     NodeType `json:"type,omitempty"`
	Data     NodeData `json:"data,omitempty"`
	Location Location `json:"location,omitempty"`
	Parent   *Node    `json:"-"`
}

type NodeData interface {
	NodeType() NodeType
	AsNode() *Node
	MarshalJSON() ([]byte, error)
	ForEachChild(v Visitor) bool
	ContainerBaseData() *ContainerBase
}

type NodeBase struct {
	Node `json:"-"`
}

func NewNode[T NodeData](nodeData T, location Location) T {
	node := nodeData.AsNode()
	node.Type = nodeData.NodeType()
	node.Location = location
	node.Data = nodeData

	node.ForEachChild(func(n *Node) bool {
		n.Parent = node
		return false
	})

	return nodeData
}

func (node *Node) AsNode() *Node { return node }

func (node *Node) AsVariableDeclaration() *VariableDeclaration {
	return node.Data.(*VariableDeclaration)
}
func (node *Node) AsFunctionDeclaration() *FunctionDeclaration {
	return node.Data.(*FunctionDeclaration)
}

func (node *Node) AsDecimalLiteral() *DecimalLiteral     { return node.Data.(*DecimalLiteral) }
func (node *Node) AsStringLiteral() *StringLiteral       { return node.Data.(*StringLiteral) }
func (node *Node) AsBooleanLiteral() *BooleanLiteral     { return node.Data.(*BooleanLiteral) }
func (node *Node) AsNullLiteral() *NullLiteral           { return node.Data.(*NullLiteral) }
func (node *Node) AsCallExpression() *CallExpression     { return node.Data.(*CallExpression) }
func (node *Node) AsIdentifier() *Identifier             { return node.Data.(*Identifier) }
func (node *Node) AsBinaryExpression() *BinaryExpression { return node.Data.(*BinaryExpression) }
func (node *Node) AsArrayExpression() *ArrayExpression   { return node.Data.(*ArrayExpression) }
func (node *Node) AsSpreadElement() *SpreadElement       { return node.Data.(*SpreadElement) }
func (node *Node) AsObjectExpression() *ObjectExpression { return node.Data.(*ObjectExpression) }
func (node *Node) AsObjectProperty() *ObjectProperty     { return node.Data.(*ObjectProperty) }
func (node *Node) AsObjectMethod() *ObjectMethod         { return node.Data.(*ObjectMethod) }
func (node *Node) AsMemberExpression() *MemberExpression { return node.Data.(*MemberExpression) }
func (node *Node) AsRestElement() *RestElement           { return node.Data.(*RestElement) }
func (node *Node) AsUpdateExpression() *UpdateExpression { return node.Data.(*UpdateExpression) }
func (node *Node) AsAssignmentExpression() *AssignmentExpression {
	return node.Data.(*AssignmentExpression)
}
func (node *Node) AsFunctionExpression() *FunctionExpression { return node.Data.(*FunctionExpression) }

func (node *Node) AsIfStatement() *IfStatement         { return node.Data.(*IfStatement) }
func (node *Node) AsForStatement() *ForStatement       { return node.Data.(*ForStatement) }
func (node *Node) AsBlockStatement() *BlockStatement   { return node.Data.(*BlockStatement) }
func (node *Node) AsSourceFile() *SourceFile           { return node.Data.(*SourceFile) }
func (node *Node) AsReturnStatement() *ReturnStatement { return node.Data.(*ReturnStatement) }
func (node *Node) AsExpressionStatement() *ExpressionStatement {
	return node.Data.(*ExpressionStatement)
}

func (node *Node) AsInterfaceDeclaration() *InterfaceDeclaration {
	return node.Data.(*InterfaceDeclaration)
}

func (node *Node) AsInterfaceBody() *InterfaceBody {
	return node.Data.(*InterfaceBody)
}

func (node *Node) AsPropertySignature() *PropertySignature {
	return node.Data.(*PropertySignature)
}

func (node *Node) AsTypeReference() *TypeReference {
	return node.Data.(*TypeReference)
}

func (node *Node) AsTypeLiteral() *TypeLiteral {
	return node.Data.(*TypeLiteral)
}

func (node *Node) AsTypeAliasDeclaration() *TypeAliasDeclaration {
	return node.Data.(*TypeAliasDeclaration)
}

func (node *Node) AsFunctionType() *FunctionType {
	return node.Data.(*FunctionType)
}

func (node *Node) AsArrayType() *ArrayType {
	return node.Data.(*ArrayType)
}

func (node *Node) AsImportDeclaration() *ImportDeclaration {
	return node.Data.(*ImportDeclaration)
}

func (node *Node) AsImportDefaultSpecifier() *ImportDefaultSpecifier {
	return node.Data.(*ImportDefaultSpecifier)
}

func (node *Node) AsImportNamespaceSpecifier() *ImportNamespaceSpecifier {
	return node.Data.(*ImportNamespaceSpecifier)
}

func (node *Node) AsImportSpecifier() *ImportSpecifier {
	return node.Data.(*ImportSpecifier)
}

func (node *Node) AsUnionType() *UnionType {
	return node.Data.(*UnionType)
}

func (node *Node) ForEachChild(v Visitor) bool     { return node.Data.ForEachChild(v) }
func (node *NodeBase) ForEachChild(v Visitor) bool { return false }

func (node *Node) ContainerBaseData() *ContainerBase     { return node.Data.ContainerBaseData() }
func (node *NodeBase) ContainerBaseData() *ContainerBase { return nil }

func (node *Node) TypeNode() *Node {
	switch node.Type {
	case NodeTypeVariableDeclaration:
		if node.AsVariableDeclaration().Identifier.TypeAnnotation != nil {
			return node.AsVariableDeclaration().Identifier.TypeAnnotation.TypeAnnotation
		}
	case NodeTypePropertySignature:
		if node.AsPropertySignature().TypeAnnotation != nil {
			return node.AsPropertySignature().TypeAnnotation.TypeAnnotation
		}
	case NodeTypeTypeAliasDeclaration:
		if node.AsTypeAliasDeclaration().TypeAnnotation != nil {
			return node.AsTypeAliasDeclaration().TypeAnnotation.TypeAnnotation
		}
	case NodeTypeFunctionDeclaration:
		if node.AsFunctionDeclaration().TypeAnnotation != nil {
			return node.AsFunctionDeclaration().TypeAnnotation.TypeAnnotation
		}
	case NodeTypeIdentifier:
		if node.AsIdentifier().TypeAnnotation != nil {
			return node.AsIdentifier().TypeAnnotation.TypeAnnotation
		}
	}

	return nil
}

func (node *Node) LocalScope() *Scope {
	data := node.ContainerBaseData()
	if data != nil {
		return data.Scope
	}
	return nil
}

func (node *Node) GetPrentContainer() *Scope {
	if node.Type == NodeTypeBlockStatement && node.Parent.Type == NodeTypeFunctionDeclaration {
		return node.Parent.AsFunctionDeclaration().Scope
	}

	switch node.Type {
	case NodeTypeSourceFile:
		return node.AsSourceFile().Scope
	case NodeTypeBlockStatement:
		return node.AsBlockStatement().Scope
	case NodeTypeFunctionDeclaration:
		return node.AsFunctionDeclaration().Scope
	default:
		return node.Parent.GetPrentContainer()
	}
}

func wrapNode[T any](n Node, data *T) interface{} {
	return struct {
		Node
		Data *T `json:"data"`
	}{
		Node: n,
		Data: data,
	}
}

type ContainerBase struct {
	Scope *Scope
}

func (node *ContainerBase) ContainerBaseData() *ContainerBase { return node }

type DeclarationBase struct {
	Symbol *Symbol
}

type ModifiersBase struct {
	modifiers *ModifierList
}

func (node *ModifiersBase) Modifiers() *ModifierList { return node.modifiers }

type ExpressionStatement struct {
	NodeBase
	Expression *Node `json:"expression,omitempty"`
}

func NewExpressionStatement(expression *Node) *ExpressionStatement {
	return &ExpressionStatement{Expression: expression}
}

func (expressionStatement *ExpressionStatement) MarshalJSON() ([]byte, error) {
	type Alias ExpressionStatement
	return json.Marshal(
		wrapNode(
			*expressionStatement.AsNode(),
			(*Alias)(expressionStatement),
		),
	)
}

func (expressionStatement *ExpressionStatement) NodeType() NodeType {
	return NodeTypeExpressionStatement
}

func (expressionStatement *ExpressionStatement) ForEachChild(v Visitor) bool {
	return visit(v, expressionStatement.Expression)
}

type VariableDeclaration struct {
	NodeBase
	DeclarationBase `json:"-"`
	ModifiersBase   `json:"-"`
	Identifier      *Identifier  `json:"identifier,omitempty"`
	Initializer     *Initializer `json:"initializer,omitempty"`
	Declare         bool         `json:"declare,omitempty"`
}

func NewVariableDeclaration(identifier *Identifier, initializer *Initializer, modifiers *ModifierList) *VariableDeclaration {
	return &VariableDeclaration{
		Identifier:  identifier,
		Initializer: initializer,
		ModifiersBase: ModifiersBase{
			modifiers: modifiers,
		},
	}
}

func (variableDeclaration *VariableDeclaration) MarshalJSON() ([]byte, error) {
	type Alias VariableDeclaration
	return json.Marshal(
		wrapNode(
			*variableDeclaration.AsNode(),
			(*Alias)(variableDeclaration),
		),
	)
}

func (variableDeclaration *VariableDeclaration) NodeType() NodeType {
	return NodeTypeVariableDeclaration
}

func (variableDeclaration *VariableDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, variableDeclaration.Identifier.AsNode()) || (variableDeclaration.Initializer != nil && visit(v, variableDeclaration.Initializer.AsNode()))
}

type TypeAnnotation struct {
	NodeBase
	TypeAnnotation *Node `json:"type_annotation,omitempty"`
}

func NewTypeAnnotation(typeAnnotation *Node) *TypeAnnotation {
	return &TypeAnnotation{TypeAnnotation: typeAnnotation}
}

func (typeAnnotation *TypeAnnotation) MarshalJSON() ([]byte, error) {
	type Alias TypeAnnotation
	return json.Marshal(
		wrapNode(
			*typeAnnotation.AsNode(),
			(*Alias)(typeAnnotation),
		),
	)
}

func (typeAnnotation *TypeAnnotation) NodeType() NodeType {
	return NodeTypeTypeAnnotation
}

func (typeAnnotation *TypeAnnotation) ForEachChild(v Visitor) bool {
	return visit(v, typeAnnotation.TypeAnnotation)
}

type StringKeyword struct {
	NodeBase
}

func NewStringKeyword() *StringKeyword {
	return &StringKeyword{}
}

func (stringKeyword *StringKeyword) MarshalJSON() ([]byte, error) {
	type Alias StringKeyword
	return json.Marshal(
		wrapNode(
			*stringKeyword.AsNode(),
			(*Alias)(stringKeyword),
		),
	)
}

func (stringKeyword *StringKeyword) NodeType() NodeType {
	return NodeTypeStringKeyword
}

type NumberKeyword struct {
	NodeBase
}

func NewNumberKeyword() *NumberKeyword {
	return &NumberKeyword{}
}

func (numberKeyword *NumberKeyword) MarshalJSON() ([]byte, error) {
	type Alias NumberKeyword
	return json.Marshal(
		wrapNode(
			*numberKeyword.AsNode(),
			(*Alias)(numberKeyword),
		),
	)
}

func (numberKeyword *NumberKeyword) NodeType() NodeType {
	return NodeTypeNumberKeyword
}

type BooleanKeyword struct {
	NodeBase
}

func NewBooleanKeyword() *BooleanKeyword {
	return &BooleanKeyword{}
}

func (booleanKeyword *BooleanKeyword) MarshalJSON() ([]byte, error) {
	type Alias BooleanKeyword
	return json.Marshal(
		wrapNode(
			*booleanKeyword.AsNode(),
			(*Alias)(booleanKeyword),
		),
	)
}

func (booleanKeyword *BooleanKeyword) NodeType() NodeType {
	return NodeTypeBooleanKeyword
}

type NullKeyword struct {
	NodeBase
}

func NewNullKeyword() *NullKeyword {
	return &NullKeyword{}
}

func (nullKeyword *NullKeyword) MarshalJSON() ([]byte, error) {
	type Alias NullKeyword
	return json.Marshal(
		wrapNode(
			*nullKeyword.AsNode(),
			(*Alias)(nullKeyword),
		),
	)
}

func (nullKeyword *NullKeyword) NodeType() NodeType {
	return NodeTypeNullKeyword
}

type AnyKeyword struct {
	NodeBase
}

func NewAnyKeyword() *AnyKeyword {
	return &AnyKeyword{}
}

func (anyKeyword *AnyKeyword) MarshalJSON() ([]byte, error) {
	type Alias AnyKeyword
	return json.Marshal(
		wrapNode(
			*anyKeyword.AsNode(),
			(*Alias)(anyKeyword),
		),
	)
}

func (anyKeyword *AnyKeyword) NodeType() NodeType {
	return NodeTypeAnyKeyword
}

type Identifier struct {
	NodeBase
	Name           string          `json:"name,omitempty"`
	OriginalName   *string         `json:"original_name,omitempty"`
	TypeAnnotation *TypeAnnotation `json:"type_annotation,omitempty"`
	Optional       bool            `json:"optional,omitempty"`
}

func NewIdentifier(name string, typeAnnotation *TypeAnnotation, optional bool) *Identifier {
	return &Identifier{Name: name, TypeAnnotation: typeAnnotation, OriginalName: nil, Optional: optional}
}

func (identifier *Identifier) MarshalJSON() ([]byte, error) {
	type Alias Identifier
	return json.Marshal(
		wrapNode(
			*identifier.AsNode(),
			(*Alias)(identifier),
		),
	)
}

func (identifier *Identifier) NodeType() NodeType {
	return NodeTypeIdentifier
}

func (identifier *Identifier) ForEachChild(v Visitor) bool {
	return identifier.TypeAnnotation != nil && visit(v, identifier.TypeAnnotation.AsNode())
}

type Initializer struct {
	NodeBase
	Expression *Node `json:"expression,omitempty"`
}

func NewInitializer(expression *Node) *Initializer {
	return &Initializer{Expression: expression}
}

func (initializer *Initializer) MarshalJSON() ([]byte, error) {
	type Alias Initializer
	return json.Marshal(
		wrapNode(
			*initializer.AsNode(),
			(*Alias)(initializer),
		),
	)
}

func (initializer *Initializer) NodeType() NodeType {
	return NodeTypeInitializer
}

func (initializer *Initializer) ForEachChild(v Visitor) bool {
	return visit(v, initializer.Expression)
}

type StringLiteral struct {
	NodeBase
	Value string `json:"value,omitempty"`
}

func NewStringLiteral(value string) *StringLiteral {
	return &StringLiteral{Value: value}
}

func (stringLiteral *StringLiteral) MarshalJSON() ([]byte, error) {
	type Alias StringLiteral
	return json.Marshal(
		wrapNode(
			*stringLiteral.AsNode(),
			(*Alias)(stringLiteral),
		),
	)
}

func (stringLiteral *StringLiteral) NodeType() NodeType {
	return NodeTypeStringLiteral
}

type NullLiteral struct {
	NodeBase
}

func NewNullLiteral() *NullLiteral {
	return &NullLiteral{}
}

func (nullLiteral *NullLiteral) MarshalJSON() ([]byte, error) {
	type Alias NullLiteral
	return json.Marshal(
		wrapNode(
			*nullLiteral.AsNode(),
			(*Alias)(nullLiteral),
		),
	)
}

func (nullLiteral *NullLiteral) NodeType() NodeType {
	return NodeTypeNullLiteral
}

type BooleanLiteral struct {
	NodeBase
	Value bool `json:"value,omitempty"`
}

func NewBooleanLiteral(value bool) *BooleanLiteral {
	return &BooleanLiteral{Value: value}
}

func (booleanLiteral *BooleanLiteral) MarshalJSON() ([]byte, error) {
	type Alias BooleanLiteral
	return json.Marshal(
		wrapNode(
			*booleanLiteral.AsNode(),
			(*Alias)(booleanLiteral),
		),
	)
}

func (booleanLiteral *BooleanLiteral) NodeType() NodeType {
	return NodeTypeBooleanLiteral
}

type DecimalLiteral struct {
	NodeBase
	Value string `json:"value,omitempty"`
}

func NewDecimalLiteral(value string) *DecimalLiteral {
	return &DecimalLiteral{Value: value}
}

func (decimalLiteral *DecimalLiteral) MarshalJSON() ([]byte, error) {
	type Alias DecimalLiteral
	return json.Marshal(
		wrapNode(
			*decimalLiteral.AsNode(),
			(*Alias)(decimalLiteral),
		),
	)
}

func (decimalLiteral *DecimalLiteral) NodeType() NodeType {
	return NodeTypeDecimalLiteral
}

type IfStatement struct {
	NodeBase
	TestExpression      *Node `json:"test_expression,omitempty"`
	ConsequentStatement *Node `json:"consequence_statement,omitempty"`
	AlternateStatement  *Node `json:"alternate_statement,omitempty"`
}

func NewIfStatement(test *Node, consequent *Node, alternate *Node) *IfStatement {
	return &IfStatement{TestExpression: test, ConsequentStatement: consequent, AlternateStatement: alternate}
}

func (ifStatement *IfStatement) MarshalJSON() ([]byte, error) {
	type Alias IfStatement
	return json.Marshal(
		wrapNode(
			*ifStatement.AsNode(),
			(*Alias)(ifStatement),
		),
	)
}

func (ifStatement *IfStatement) NodeType() NodeType {
	return NodeTypeIfStatement
}

func (ifStatement *IfStatement) ForEachChild(v Visitor) bool {
	return visit(v, ifStatement.TestExpression) || visit(v, ifStatement.ConsequentStatement) || visit(v, ifStatement.AlternateStatement)
}

type BlockStatement struct {
	NodeBase
	ContainerBase `json:"-"`
	Body          []*Node `json:"body,omitempty"`
}

func NewBlockStatement(body []*Node) *BlockStatement {
	return &BlockStatement{Body: body}
}

func (blockStatement *BlockStatement) MarshalJSON() ([]byte, error) {
	type Alias BlockStatement
	return json.Marshal(
		wrapNode(
			*blockStatement.AsNode(),
			(*Alias)(blockStatement),
		),
	)
}

func (blockStatement *BlockStatement) NodeType() NodeType {
	return NodeTypeBlockStatement
}

func (blockStatement *BlockStatement) ForEachChild(v Visitor) bool {
	return visitNodes(v, blockStatement.Body)
}

func (blockStatement *BlockStatement) ContainerBaseData() *ContainerBase {
	return &blockStatement.ContainerBase
}

type AssignmentExpression struct {
	NodeBase
	Operator string `json:"operator,omitempty"`
	Left     *Node  `json:"left,omitempty"`
	Right    *Node  `json:"right,omitempty"`
}

func NewAssignmentExpression(operator string, left *Node, right *Node) *AssignmentExpression {
	return &AssignmentExpression{Operator: operator, Left: left, Right: right}
}

func (assignmentExpression *AssignmentExpression) MarshalJSON() ([]byte, error) {
	type Alias AssignmentExpression
	return json.Marshal(
		wrapNode(
			*assignmentExpression.AsNode(),
			(*Alias)(assignmentExpression),
		),
	)
}

func (assignmentExpression *AssignmentExpression) NodeType() NodeType {
	return NodeTypeAssignmentExpression
}

func (assignmentExpression *AssignmentExpression) ForEachChild(v Visitor) bool {
	return visit(v, assignmentExpression.Left) || visit(v, assignmentExpression.Right)
}

type FunctionDeclaration struct {
	NodeBase
	ContainerBase   `json:"-"`
	DeclarationBase `json:"-"`
	ModifiersBase   `json:"-"`
	ID              *Identifier                `json:"id,omitempty"`
	TypeParameters  *TypeParametersDeclaration `json:"type_parameters,omitempty"`
	Params          []*Node                    `json:"params,omitempty"`
	Body            *BlockStatement            `json:"body,omitempty"`
	TypeAnnotation  *TypeAnnotation            `json:"type_annotation,omitempty"`
}

func NewFunctionDeclaration(
	id *Identifier,
	typeParameters *TypeParametersDeclaration,
	params []*Node,
	body *BlockStatement,
	typeAnnotation *TypeAnnotation,
	modifiers *ModifierList,
) *FunctionDeclaration {
	return &FunctionDeclaration{
		ID:             id,
		TypeParameters: typeParameters,
		Params:         params,
		Body:           body,
		TypeAnnotation: typeAnnotation,
		ModifiersBase: ModifiersBase{
			modifiers: modifiers,
		},
	}
}

func (functionDeclaration *FunctionDeclaration) MarshalJSON() ([]byte, error) {
	type Alias FunctionDeclaration
	return json.Marshal(
		wrapNode(
			*functionDeclaration.AsNode(),
			(*Alias)(functionDeclaration),
		),
	)
}

func (functionDeclaration *FunctionDeclaration) NodeType() NodeType {
	return NodeTypeFunctionDeclaration
}

func (functionDeclaration *FunctionDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, functionDeclaration.ID.AsNode()) ||
		visitNodes(v, functionDeclaration.Params) ||
		(functionDeclaration.Body != nil && visit(v, functionDeclaration.Body.AsNode())) ||
		(functionDeclaration.TypeAnnotation != nil && visit(v, functionDeclaration.TypeAnnotation.AsNode()))
}

func (functionDeclaration *FunctionDeclaration) ContainerBaseData() *ContainerBase {
	return &functionDeclaration.ContainerBase
}

type CallExpression struct {
	NodeBase
	Callee         *Node                       `json:"callee,omitempty"`
	TypeParameters *TypeParameterInstantiation `json:"type_parameters,omitempty"`
	Args           []*Node                     `json:"args,omitempty"`
}

func NewCallExpression(callee *Node, typeParameters *TypeParameterInstantiation, args []*Node) *CallExpression {
	return &CallExpression{Callee: callee, TypeParameters: typeParameters, Args: args}
}

func (callExpression *CallExpression) MarshalJSON() ([]byte, error) {
	type Alias CallExpression
	return json.Marshal(
		wrapNode(
			*callExpression.AsNode(),
			(*Alias)(callExpression),
		),
	)
}

func (callExpression *CallExpression) NodeType() NodeType {
	return NodeTypeCallExpression
}

func (callExpression *CallExpression) ForEachChild(v Visitor) bool {
	return visit(v, callExpression.Callee) ||
		(callExpression.TypeParameters != nil && visit(v, callExpression.TypeParameters.AsNode())) ||
		visitNodes(v, callExpression.Args)
}

type MemberExpression struct {
	NodeBase
	Object   *Node `json:"object,omitempty"`
	Property *Node `json:"property,omitempty"`
	Computed bool
}

func NewMemberExpression(object *Node, property *Node, computed bool) *MemberExpression {
	return &MemberExpression{Object: object, Property: property, Computed: computed}
}

func (memberExpression *MemberExpression) MarshalJSON() ([]byte, error) {
	type Alias MemberExpression
	return json.Marshal(
		wrapNode(
			*memberExpression.AsNode(),
			(*Alias)(memberExpression),
		),
	)
}

func (memberExpression *MemberExpression) NodeType() NodeType {
	return NodeTypeMemberExpression
}

func (memberExpression *MemberExpression) ForEachChild(v Visitor) bool {
	return visit(v, memberExpression.Object) || visit(v, memberExpression.Property)
}

func (memberExpression *MemberExpression) PropertyName() string {
	switch memberExpression.Property.Type {
	case NodeTypeIdentifier:
		return memberExpression.Property.AsIdentifier().Name
	default:
		return ""
	}
}

type ImportSpecifier struct {
	NodeBase
	Local    *Identifier `json:"local,omitempty"`
	Imported *Node       `json:"imported,omitempty"`
}

func NewImportSpecifier(local *Identifier, imported *Node) *ImportSpecifier {
	return &ImportSpecifier{Local: local, Imported: imported}
}

func (importSpecifier *ImportSpecifier) MarshalJSON() ([]byte, error) {
	type Alias ImportSpecifier
	return json.Marshal(
		wrapNode(
			*importSpecifier.AsNode(),
			(*Alias)(importSpecifier),
		),
	)
}

func (importSpecifier *ImportSpecifier) NodeType() NodeType {
	return NodeTypeImportSpecifier
}

func (importSpecifier *ImportSpecifier) ForEachChild(v Visitor) bool {
	return visit(v, importSpecifier.Local.AsNode()) || visit(v, importSpecifier.Imported)
}

type ImportDefaultSpecifier struct {
	NodeBase
	Local *Identifier `json:"local,omitempty"`
}

func NewImportDefaultSpecifier(local *Identifier) *ImportDefaultSpecifier {
	return &ImportDefaultSpecifier{Local: local}
}

func (importDefaultSpecifier *ImportDefaultSpecifier) MarshalJSON() ([]byte, error) {
	type Alias ImportDefaultSpecifier
	return json.Marshal(
		wrapNode(
			*importDefaultSpecifier.AsNode(),
			(*Alias)(importDefaultSpecifier),
		),
	)
}

func (importDefaultSpecifier *ImportDefaultSpecifier) NodeType() NodeType {
	return NodeTypeImportDefaultSpecifier
}

func (importDefaultSpecifier *ImportDefaultSpecifier) ForEachChild(v Visitor) bool {
	return visit(v, importDefaultSpecifier.Local.AsNode())
}

type ImportNamespaceSpecifier struct {
	NodeBase
	Local *Identifier `json:"local,omitempty"`
}

func NewImportNamespaceSpecifier(local *Identifier) *ImportNamespaceSpecifier {
	return &ImportNamespaceSpecifier{Local: local}
}

func (importNamespaceSpecifier *ImportNamespaceSpecifier) MarshalJSON() ([]byte, error) {
	type Alias ImportNamespaceSpecifier
	return json.Marshal(
		wrapNode(
			*importNamespaceSpecifier.AsNode(),
			(*Alias)(importNamespaceSpecifier),
		),
	)
}

func (importNamespaceSpecifier *ImportNamespaceSpecifier) NodeType() NodeType {
	return NodeTypeImportNamespaceSpecifier
}

func (importNamespaceSpecifier *ImportNamespaceSpecifier) ForEachChild(v Visitor) bool {
	return visit(v, importNamespaceSpecifier.Local.AsNode())
}

type ImportDeclaration struct {
	NodeBase
	Specifiers []*Node        `json:"specifiers,omitempty"`
	Source     *StringLiteral `json:"source,omitempty"`
}

func NewImportDeclaration(specifiers []*Node, source *StringLiteral) *ImportDeclaration {
	return &ImportDeclaration{Specifiers: specifiers, Source: source}
}

func (importDeclaration *ImportDeclaration) MarshalJSON() ([]byte, error) {
	type Alias ImportDeclaration
	return json.Marshal(
		wrapNode(
			*importDeclaration.AsNode(),
			(*Alias)(importDeclaration),
		),
	)
}

func (importDeclaration *ImportDeclaration) NodeType() NodeType {
	return NodeTypeImportDeclaration
}

func (importDeclaration *ImportDeclaration) ForEachChild(v Visitor) bool {
	return visitNodes(v, importDeclaration.Specifiers) || visit(v, importDeclaration.Source.AsNode())
}

type ExportNamedDeclaration struct {
	NodeBase
	Declaration *Node              `json:"declaration,omitempty"`
	Specifiers  []*ExportSpecifier `json:"specifiers,omitempty"`
	Source      *StringLiteral     `json:"source,omitempty"`
}

func NewExportNamedDeclaration(declaration *Node, specifiers []*ExportSpecifier, source *StringLiteral) *ExportNamedDeclaration {
	return &ExportNamedDeclaration{Declaration: declaration, Specifiers: specifiers, Source: source}
}

func (exportNamedDeclaration *ExportNamedDeclaration) MarshalJSON() ([]byte, error) {
	type Alias ExportNamedDeclaration
	return json.Marshal(
		wrapNode(
			*exportNamedDeclaration.AsNode(),
			(*Alias)(exportNamedDeclaration),
		),
	)
}

func (exportNamedDeclaration *ExportNamedDeclaration) NodeType() NodeType {
	return NodeTypeExportNamedDeclaration
}

func (exportNamedDeclaration *ExportNamedDeclaration) ForEachChild(v Visitor) bool {
	_specifiers := []*Node{}
	for _, specifier := range exportNamedDeclaration.Specifiers {
		_specifiers = append(_specifiers, specifier.AsNode())
	}

	return visit(v, exportNamedDeclaration.Declaration) || visitNodes(v, _specifiers) || visit(v, exportNamedDeclaration.Source.AsNode())
}

type ExportDefaultDeclaration struct {
	NodeBase
	Declaration *Node `json:"declaration,omitempty"`
}

func NewExportDefaultDeclaration(declaration *Node) *ExportDefaultDeclaration {
	return &ExportDefaultDeclaration{Declaration: declaration}
}

func (exportDefaultDeclaration *ExportDefaultDeclaration) MarshalJSON() ([]byte, error) {
	type Alias ExportDefaultDeclaration
	return json.Marshal(
		wrapNode(
			*exportDefaultDeclaration.AsNode(),
			(*Alias)(exportDefaultDeclaration),
		),
	)
}

func (exportDefaultDeclaration *ExportDefaultDeclaration) NodeType() NodeType {
	return NodeTypeExportDefaultDeclaration
}

func (exportDefaultDeclaration *ExportDefaultDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, exportDefaultDeclaration.Declaration)
}

type ExportSpecifier struct {
	NodeBase
	Local    *Identifier `json:"local,omitempty"`
	Exported *Node       `json:"exported,omitempty"`
}

func NewExportSpecifier(local *Identifier, exported *Node) *ExportSpecifier {
	return &ExportSpecifier{Local: local, Exported: exported}
}

func (exportSpecifier *ExportSpecifier) MarshalJSON() ([]byte, error) {
	type Alias ExportSpecifier
	return json.Marshal(
		wrapNode(
			*exportSpecifier.AsNode(),
			(*Alias)(exportSpecifier),
		),
	)
}

func (exportSpecifier *ExportSpecifier) NodeType() NodeType {
	return NodeTypeExportSpecifier
}

func (exportSpecifier *ExportSpecifier) ForEachChild(v Visitor) bool {
	return visit(v, exportSpecifier.Local.AsNode()) || visit(v, exportSpecifier.Exported)
}

type BinaryExpression struct {
	NodeBase
	Operator BinaryExpressionOperator `json:"operator,omitempty"`
	Left     *Node                    `json:"left,omitempty"`
	Right    *Node                    `json:"right,omitempty"`
}

func NewBinaryExpression(operator BinaryExpressionOperator, left *Node, right *Node) *BinaryExpression {
	return &BinaryExpression{Operator: operator, Left: left, Right: right}
}

func (binaryExpression *BinaryExpression) MarshalJSON() ([]byte, error) {
	type Alias BinaryExpression
	return json.Marshal(
		wrapNode(
			*binaryExpression.AsNode(),
			(*Alias)(binaryExpression),
		),
	)
}

func (binaryExpression *BinaryExpression) NodeType() NodeType {
	return NodeTypeBinaryExpression
}

func (binaryExpression *BinaryExpression) ForEachChild(v Visitor) bool {
	return visit(v, binaryExpression.Left) || visit(v, binaryExpression.Right)
}

type SpreadElement struct {
	NodeBase
	Argument *Node `json:"argument,omitempty"`
}

func NewSpreadElement(argument *Node) *SpreadElement {
	return &SpreadElement{Argument: argument}
}

func (spreadElement *SpreadElement) MarshalJSON() ([]byte, error) {
	type Alias SpreadElement
	return json.Marshal(
		wrapNode(
			*spreadElement.AsNode(),
			(*Alias)(spreadElement),
		),
	)
}

func (spreadElement *SpreadElement) NodeType() NodeType {
	return NodeTypeSpreadElement
}

func (spreadElement *SpreadElement) ForEachChild(v Visitor) bool {
	return visit(v, spreadElement.Argument)
}

type ArrayExpression struct {
	NodeBase
	Elements []*Node `json:"elements,omitempty"`
}

func NewArrayExpression(elements []*Node) *ArrayExpression {
	return &ArrayExpression{Elements: elements}
}

func (arrayExpression *ArrayExpression) MarshalJSON() ([]byte, error) {
	type Alias ArrayExpression
	return json.Marshal(
		wrapNode(
			*arrayExpression.AsNode(),
			(*Alias)(arrayExpression),
		),
	)
}

func (arrayExpression *ArrayExpression) NodeType() NodeType {
	return NodeTypeArrayExpression
}

func (arrayExpression *ArrayExpression) ForEachChild(v Visitor) bool {
	return visitNodes(v, arrayExpression.Elements)
}

type ObjectProperty struct {
	NodeBase
	Key   *Node `json:"key,omitempty"`
	Value *Node `json:"value,omitempty"`
}

func NewObjectProperty(key *Node, value *Node) *ObjectProperty {
	return &ObjectProperty{Key: key, Value: value}
}

func (objectProperty *ObjectProperty) MarshalJSON() ([]byte, error) {
	type Alias ObjectProperty
	return json.Marshal(
		wrapNode(
			*objectProperty.AsNode(),
			(*Alias)(objectProperty),
		),
	)
}

func (objectProperty *ObjectProperty) NodeType() NodeType {
	return NodeTypeObjectProperty
}

func (objectProperty *ObjectProperty) ForEachChild(v Visitor) bool {
	return visit(v, objectProperty.Key) || visit(v, objectProperty.Value)
}

func (objectProperty *ObjectProperty) Name() string {
	node := objectProperty.Key

	switch node.Type {
	case NodeTypeStringLiteral:
		return node.AsStringLiteral().Value
	case NodeTypeIdentifier:
		return node.AsIdentifier().Name
	case NodeTypeDecimalLiteral:
		return node.AsDecimalLiteral().Value
	}

	return ""
}

type ObjectMethod struct {
	NodeBase
	Kind   string          `json:"kind,omitempty"`
	Key    *Node           `json:"key,omitempty"`
	Params []*Identifier   `json:"Params,omitempty"`
	Body   *BlockStatement `json:"body,omitempty"`
}

func NewObjectMethod(kind string, key *Node, params []*Identifier, body *BlockStatement) *ObjectMethod {
	return &ObjectMethod{Kind: kind, Key: key, Params: params, Body: body}
}

func (objectMethod *ObjectMethod) MarshalJSON() ([]byte, error) {
	type Alias ObjectMethod
	return json.Marshal(
		wrapNode(
			*objectMethod.AsNode(),
			(*Alias)(objectMethod),
		),
	)
}

func (objectMethod *ObjectMethod) NodeType() NodeType {
	return NodeTypeObjectMethod
}

func (objectMethod *ObjectMethod) ForEachChild(v Visitor) bool {
	params := []*Node{}
	for _, param := range objectMethod.Params {
		params = append(params, param.AsNode())
	}

	return visit(v, objectMethod.Key) || visitNodes(v, params) || visit(v, objectMethod.Body.AsNode())
}

type ObjectExpression struct {
	NodeBase
	Properties []*Node `json:"properties,omitempty"`
}

func NewObjectExpression(properties []*Node) *ObjectExpression {
	return &ObjectExpression{Properties: properties}
}

func (objectExpression *ObjectExpression) MarshalJSON() ([]byte, error) {
	type Alias ObjectExpression
	return json.Marshal(
		wrapNode(
			*objectExpression.AsNode(),
			(*Alias)(objectExpression),
		),
	)
}

func (objectExpression *ObjectExpression) NodeType() NodeType {
	return NodeTypeObjectExpression
}

func (objectExpression *ObjectExpression) ForEachChild(v Visitor) bool {
	return visitNodes(v, objectExpression.Properties)
}

type DirectiveLiteral struct {
	NodeBase
	Value string `json:"value,omitempty"`
}

func NewDirectiveLiteral(value string) *DirectiveLiteral {
	return &DirectiveLiteral{Value: value}
}

func (directiveLiteral *DirectiveLiteral) MarshalJSON() ([]byte, error) {
	type Alias DirectiveLiteral
	return json.Marshal(
		wrapNode(
			*directiveLiteral.AsNode(),
			(*Alias)(directiveLiteral),
		),
	)
}

func (directiveLiteral *DirectiveLiteral) NodeType() NodeType {
	return NodeTypeDirectiveLiteral
}

type Directive struct {
	NodeBase
	Value *DirectiveLiteral `json:"value,omitempty"`
}

func NewDirective(value *DirectiveLiteral) *Directive {
	return &Directive{Value: value}
}

func (directive *Directive) MarshalJSON() ([]byte, error) {
	type Alias Directive
	return json.Marshal(
		wrapNode(
			*directive.AsNode(),
			(*Alias)(directive),
		),
	)
}

func (directive *Directive) NodeType() NodeType {
	return NodeTypeDirective
}

func (directive *Directive) ForEachChild(v Visitor) bool {
	return visit(v, directive.Value.AsNode())
}

type SourceFile struct {
	NodeBase
	ContainerBase `json:"-"`
	Name          string       `json:"name,omitempty"`
	Body          []*Node      `json:"body,omitempty"`
	Directives    []*Directive `json:"directives,omitempty"`

	ExternalModuleIndicator *Node  `json:"-"`
	Path                    string `json:"-"`
	IsDeclarationFile       bool   `json:"-"`
}

func NewSourceFile(body []*Node, directives []*Directive, isDeclarationFile bool) *SourceFile {
	return &SourceFile{Name: "", Body: body, Directives: directives, IsDeclarationFile: isDeclarationFile}
}

func (sourceFile *SourceFile) MarshalJSON() ([]byte, error) {
	type Alias SourceFile
	return json.Marshal(
		wrapNode(
			*sourceFile.AsNode(),
			(*Alias)(sourceFile),
		),
	)
}

func (sourceFile *SourceFile) NodeType() NodeType {
	return NodeTypeSourceFile
}

func (sourceFile *SourceFile) ForEachChild(v Visitor) bool {
	directives := []*Node{}
	for _, directive := range sourceFile.Directives {
		directives = append(directives, directive.AsNode())
	}

	return visitNodes(v, directives) || visitNodes(v, sourceFile.Body) || visit(v, sourceFile.ExternalModuleIndicator)
}

func (sourceFile *SourceFile) ContainerBaseData() *ContainerBase {
	return &sourceFile.ContainerBase
}

type InterfaceDeclaration struct {
	NodeBase
	DeclarationBase `json:"-"`
	Id              *Identifier                `json:"id,omitempty"`
	TypeParameters  *TypeParametersDeclaration `json:"type_parameters,omitempty"`
	Body            *InterfaceBody             `json:"body,omitempty"`
	Extends         []*Node                    `json:"extends,omitempty"`
}

func NewInterfaceDeclaration(id *Identifier, typeParameters *TypeParametersDeclaration, body *InterfaceBody, extends []*Node) *InterfaceDeclaration {
	return &InterfaceDeclaration{
		Id:             id,
		TypeParameters: typeParameters,
		Body:           body,
		Extends:        extends,
	}
}

func (interfaceDeclaration *InterfaceDeclaration) MarshalJSON() ([]byte, error) {
	type Alias InterfaceDeclaration
	return json.Marshal(
		wrapNode(
			*interfaceDeclaration.AsNode(),
			(*Alias)(interfaceDeclaration),
		),
	)
}

func (interfaceDeclaration *InterfaceDeclaration) NodeType() NodeType {
	return NodeTypeInterfaceDeclaration
}

func (interfaceDeclaration *InterfaceDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, interfaceDeclaration.Id.AsNode()) || visit(v, interfaceDeclaration.Body.AsNode())
}

type InterfaceBody struct {
	NodeBase
	Body []*Node `json:"body,omitempty"`
}

func NewInterfaceBody(body []*Node) *InterfaceBody {
	return &InterfaceBody{Body: body}
}

func (interfaceBody *InterfaceBody) MarshalJSON() ([]byte, error) {
	type Alias InterfaceBody
	return json.Marshal(
		wrapNode(
			*interfaceBody.AsNode(),
			(*Alias)(interfaceBody),
		),
	)
}

func (interfaceBody *InterfaceBody) NodeType() NodeType {
	return NodeTypeInterfaceBody
}

func (interfaceBody *InterfaceBody) ForEachChild(v Visitor) bool {
	return visitNodes(v, interfaceBody.Body)
}

type PropertySignature struct {
	NodeBase
	ModifiersBase
	Key            *Node           `json:"key,omitempty"`
	TypeAnnotation *TypeAnnotation `json:"type_annotation,omitempty"`
}

func NewPropertySignature(key *Node, typeAnnotation *TypeAnnotation, modifiers *ModifierList) *PropertySignature {
	return &PropertySignature{
		Key:            key,
		TypeAnnotation: typeAnnotation,
		ModifiersBase:  ModifiersBase{modifiers: modifiers},
	}
}

func (propertySignature *PropertySignature) MarshalJSON() ([]byte, error) {
	type Alias PropertySignature
	return json.Marshal(
		wrapNode(
			*propertySignature.AsNode(),
			(*Alias)(propertySignature),
		),
	)
}

func (propertySignature *PropertySignature) NodeType() NodeType {
	return NodeTypePropertySignature
}

func (propertySignature *PropertySignature) ForEachChild(v Visitor) bool {
	return visit(v, propertySignature.Key) || visit(v, propertySignature.TypeAnnotation.AsNode())
}

type ReturnStatement struct {
	NodeBase
	Argument *Node `json:"argument,omitempty"`
}

func NewReturnStatement(argument *Node) *ReturnStatement {
	return &ReturnStatement{Argument: argument}
}

func (returnStatement *ReturnStatement) MarshalJSON() ([]byte, error) {
	type Alias ReturnStatement
	return json.Marshal(
		wrapNode(
			*returnStatement.AsNode(),
			(*Alias)(returnStatement),
		),
	)
}

func (returnStatement *ReturnStatement) NodeType() NodeType {
	return NodeTypeReturnStatement
}

func (returnStatement *ReturnStatement) ForEachChild(v Visitor) bool {
	return visit(v, returnStatement.Argument)
}

type FunctionType struct {
	NodeBase
	Params         []*Node         `json:"params,omitempty"`
	TypeAnnotation *TypeAnnotation `json:"type_annotation,omitempty"`
}

func NewFunctionType(params []*Node, typeAnnotation *TypeAnnotation) *FunctionType {
	return &FunctionType{
		Params:         params,
		TypeAnnotation: typeAnnotation,
	}
}

func (functionType *FunctionType) MarshalJSON() ([]byte, error) {
	type Alias FunctionType
	return json.Marshal(
		wrapNode(
			*functionType.AsNode(),
			(*Alias)(functionType),
		),
	)
}

func (functionType *FunctionType) NodeType() NodeType {
	return NodeTypeFunctionType
}

func (functionType *FunctionType) ForEachChild(v Visitor) bool {
	return visitNodes(v, functionType.Params) || visit(v, functionType.TypeAnnotation.AsNode())
}

type TypeAliasDeclaration struct {
	NodeBase
	DeclarationBase `json:"-"`
	Id              *Identifier                `json:"id,omitempty"`
	TypeParameters  *TypeParametersDeclaration `json:"type_parameters,omitempty"`
	TypeAnnotation  *TypeAnnotation            `json:"type_annotation,omitempty"`
}

func NewTypeAliasDeclaration(id *Identifier, typeParameters *TypeParametersDeclaration, typeAnnotation *TypeAnnotation) *TypeAliasDeclaration {
	return &TypeAliasDeclaration{
		Id:             id,
		TypeParameters: typeParameters,
		TypeAnnotation: typeAnnotation,
	}
}

func (typeAliasDeclaration *TypeAliasDeclaration) MarshalJSON() ([]byte, error) {
	type Alias TypeAliasDeclaration
	return json.Marshal(
		wrapNode(
			*typeAliasDeclaration.AsNode(),
			(*Alias)(typeAliasDeclaration),
		),
	)
}

func (typeAliasDeclaration *TypeAliasDeclaration) NodeType() NodeType {
	return NodeTypeTypeAliasDeclaration
}

func (typeAliasDeclaration *TypeAliasDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, typeAliasDeclaration.Id.AsNode()) || visit(v, typeAliasDeclaration.TypeAnnotation.AsNode())
}

type TypeLiteral struct {
	NodeBase
	Members []*Node `json:"members,omitempty"`
}

func NewTypeLiteral(members []*Node) *TypeLiteral {
	return &TypeLiteral{Members: members}
}

func (typeLiteral *TypeLiteral) MarshalJSON() ([]byte, error) {
	type Alias TypeLiteral
	return json.Marshal(
		wrapNode(
			*typeLiteral.AsNode(),
			(*Alias)(typeLiteral),
		),
	)
}

func (typeLiteral *TypeLiteral) NodeType() NodeType {
	return NodeTypeTypeLiteral
}

func (typeLiteral *TypeLiteral) ForEachChild(v Visitor) bool {
	return visitNodes(v, typeLiteral.Members)
}

type TypeReference struct {
	NodeBase
	TypeName       *Identifier                 `json:"type_name,omitempty"`
	TypeParameters *TypeParameterInstantiation `json:"type_parameters,omitempty"`
}

func NewTypeReference(TypeName *Identifier, typeParameters *TypeParameterInstantiation) *TypeReference {
	return &TypeReference{TypeName: TypeName, TypeParameters: typeParameters}
}

func (typeReference *TypeReference) MarshalJSON() ([]byte, error) {
	type Alias TypeReference
	return json.Marshal(
		wrapNode(
			*typeReference.AsNode(),
			(*Alias)(typeReference),
		),
	)
}

func (typeReference *TypeReference) NodeType() NodeType {
	return NodeTypeTypeReference
}

func (typeReference *TypeReference) ForEachChild(v Visitor) bool {
	return visit(v, typeReference.TypeName.AsNode())
}

type ArrayType struct {
	NodeBase
	ElementType *Node `json:"element_type,omitempty"`
}

func NewArrayType(elementType *Node) *ArrayType {
	return &ArrayType{ElementType: elementType}
}

func (arrayType *ArrayType) MarshalJSON() ([]byte, error) {
	type Alias ArrayType
	return json.Marshal(
		wrapNode(
			*arrayType.AsNode(),
			(*Alias)(arrayType),
		),
	)
}

func (arrayType *ArrayType) NodeType() NodeType {
	return NodeTypeArrayType
}

func (arrayType *ArrayType) ForEachChild(v Visitor) bool {
	return visit(v, arrayType.AsNode())
}

type RestElement struct {
	NodeBase
	Argument       *Identifier     `json:"argument,omitempty"`
	TypeAnnotation *TypeAnnotation `json:"type_annotation,omitempty"`
}

func NewRestElement(argument *Identifier, typeAnnotation *TypeAnnotation) *RestElement {
	return &RestElement{
		Argument:       argument,
		TypeAnnotation: typeAnnotation,
	}
}

func (restElement *RestElement) MarshalJSON() ([]byte, error) {
	type Alias RestElement
	return json.Marshal(
		wrapNode(
			*restElement.AsNode(),
			(*Alias)(restElement),
		),
	)
}

func (restElement *RestElement) NodeType() NodeType {
	return NodeTypeRestElement
}

func (restElement *RestElement) ForEachChild(v Visitor) bool {
	return visit(v, restElement.Argument.AsNode()) || (restElement.TypeAnnotation != nil && visit(v, restElement.TypeAnnotation.AsNode()))
}

type ForStatement struct {
	NodeBase
	ContainerBase `json:"-"`
	Init          *Node `json:"init,omitempty"`
	Test          *Node `json:"test,omitempty"`
	Update        *Node `json:"update,omitempty"`
	Body          *Node `json:"body,omitempty"`
}

func NewForStatement(init *Node, test *Node, update *Node, body *Node) *ForStatement {
	return &ForStatement{
		Init:   init,
		Test:   test,
		Update: update,
		Body:   body,
	}
}

func (forStatement *ForStatement) MarshalJSON() ([]byte, error) {
	type Alias ForStatement
	return json.Marshal(
		wrapNode(
			*forStatement.AsNode(),
			(*Alias)(forStatement),
		),
	)
}

func (forStatement *ForStatement) NodeType() NodeType {
	return NodeTypeForStatement
}

func (forStatement *ForStatement) ForEachChild(v Visitor) bool {
	return visit(v, forStatement.Init) ||
		visit(v, forStatement.Test) ||
		visit(v, forStatement.Update) ||
		visit(v, forStatement.Body)
}

func (forStatement *ForStatement) ContainerBaseData() *ContainerBase {
	return &forStatement.ContainerBase
}

type UpdateExpression struct {
	NodeBase
	Operator UpdateExpressionOperator `json:"operator,omitempty"`
	Argument *Node                    `json:"argument,omitempty"`
	Prefix   bool                     `json:"perfix,omitempty"`
}

func NewUpdateExpression(operator UpdateExpressionOperator, argument *Node, prefix bool) *UpdateExpression {
	return &UpdateExpression{
		Operator: operator,
		Argument: argument,
		Prefix:   prefix,
	}
}

func (updateExpression *UpdateExpression) MarshalJSON() ([]byte, error) {
	type Alias UpdateExpression
	return json.Marshal(
		wrapNode(
			*updateExpression.AsNode(),
			(*Alias)(updateExpression),
		),
	)
}

func (updateExpression *UpdateExpression) NodeType() NodeType {
	return NodeTypeUpdateExpression
}

func (updateExpression *UpdateExpression) ForEachChild(v Visitor) bool {
	return visit(v, updateExpression.Argument)
}

type ModuleDeclaration struct {
	NodeBase
	DeclarationBase `json:"-"`
	Id              *StringLiteral
	OriginalName    *string `json:"original_name,omitempty"`
	Body            *ModuleBlock
}

func NewModuleDeclaration(id *StringLiteral, body *ModuleBlock) *ModuleDeclaration {
	return &ModuleDeclaration{
		Id:   id,
		Body: body,
	}
}

func (moduleDeclaration *ModuleDeclaration) MarshalJSON() ([]byte, error) {
	type Alias ModuleDeclaration
	return json.Marshal(
		wrapNode(
			*moduleDeclaration.AsNode(),
			(*Alias)(moduleDeclaration),
		),
	)
}

func (moduleDeclaration *ModuleDeclaration) NodeType() NodeType {
	return NodeTypeModuleDeclaration
}

func (moduleDeclaration *ModuleDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, moduleDeclaration.Id.AsNode()) || visit(v, moduleDeclaration.Body.AsNode())
}

type ModuleBlock struct {
	NodeBase
	Body []*Node
}

func NewModuleBlock(body []*Node) *ModuleBlock {
	return &ModuleBlock{
		Body: body,
	}
}

func (moduleBlock *ModuleBlock) MarshalJSON() ([]byte, error) {
	type Alias ModuleBlock
	return json.Marshal(
		wrapNode(
			*moduleBlock.AsNode(),
			(*Alias)(moduleBlock),
		),
	)
}

func (moduleBlock *ModuleBlock) NodeType() NodeType {
	return NodeTypeModuleBlock
}

func (moduleBlock *ModuleBlock) ForEachChild(v Visitor) bool {
	return visitNodes(v, moduleBlock.Body)
}

type UnionType struct {
	NodeBase
	Types []*Node
}

func NewUnionType(types []*Node) *UnionType {
	return &UnionType{
		Types: types,
	}
}

func (unionType *UnionType) MarshalJSON() ([]byte, error) {
	type Alias UnionType
	return json.Marshal(
		wrapNode(
			*unionType.AsNode(),
			(*Alias)(unionType),
		),
	)
}

func (unionType *UnionType) NodeType() NodeType {
	return NodeTypeUnionType
}

func (unionType *UnionType) ForEachChild(v Visitor) bool {
	return visitNodes(v, unionType.Types)
}

type TypeParameter struct {
	NodeBase
	DeclarationBase `json:"-"`
	Name            string `json:"name,omitempty"`
}

func NewTypeParameter(name string) *TypeParameter {
	return &TypeParameter{Name: name}
}

func (typeParameter *TypeParameter) MarshalJSON() ([]byte, error) {
	type Alias TypeParameter
	return json.Marshal(
		wrapNode(
			*typeParameter.AsNode(),
			(*Alias)(typeParameter),
		),
	)
}

func (typeParameter *TypeParameter) NodeType() NodeType {
	return NodeTypeTypeParameter
}

type TypeParametersDeclaration struct {
	NodeBase
	Params []*TypeParameter `json:"params,omitempty"`
}

func NewTypeParametersDeclaration(params []*TypeParameter) *TypeParametersDeclaration {
	return &TypeParametersDeclaration{Params: params}
}

func (typeParametersDeclaration *TypeParametersDeclaration) MarshalJSON() ([]byte, error) {
	type Alias TypeParametersDeclaration
	return json.Marshal(
		wrapNode(
			*typeParametersDeclaration.AsNode(),
			(*Alias)(typeParametersDeclaration),
		),
	)
}

func (typeParametersDeclaration *TypeParametersDeclaration) NodeType() NodeType {
	return NodeTypeTypeParametersDeclaration
}

func (typeParametersDeclaration *TypeParametersDeclaration) ForEachChild(v Visitor) bool {
	_params := []*Node{}
	for _, param := range typeParametersDeclaration.Params {
		_params = append(_params, param.AsNode())
	}
	return visitNodes(v, _params)
}

type TypeParameterInstantiation struct {
	NodeBase
	Params []*Node `json:"params,omitempty"`
}

func NewTypeParameterInstantiation(params []*Node) *TypeParameterInstantiation {
	return &TypeParameterInstantiation{Params: params}
}

func (typeParameterInstantiation *TypeParameterInstantiation) MarshalJSON() ([]byte, error) {
	type Alias TypeParameterInstantiation
	return json.Marshal(
		wrapNode(
			*typeParameterInstantiation.AsNode(),
			(*Alias)(typeParameterInstantiation),
		),
	)
}

func (typeParameterInstantiation *TypeParameterInstantiation) NodeType() NodeType {
	return NodeTypeTypeParameterInstantiation
}

func (typeParameterInstantiation *TypeParameterInstantiation) ForEachChild(v Visitor) bool {
	return visitNodes(v, typeParameterInstantiation.Params)
}

type Visitor func(*Node) bool

func visit(v Visitor, node *Node) bool {
	if node != nil {
		return v(node)
	}
	return false
}

func visitNodes(v Visitor, nodes []*Node) bool {
	for _, node := range nodes {
		if v(node) {
			return true
		}
	}
	return false
}

type ThisEpxression struct {
	NodeBase
}

func NewThisEpxression() *ThisEpxression {
	return &ThisEpxression{}
}

func (thisEpxression *ThisEpxression) MarshalJSON() ([]byte, error) {
	type Alias ThisEpxression
	return json.Marshal(
		wrapNode(
			*thisEpxression.AsNode(),
			(*Alias)(thisEpxression),
		),
	)
}

func (thisEpxression *ThisEpxression) NodeType() NodeType {
	return NodeTypeThisEpxression
}

type FunctionExpression struct {
	NodeBase
	ContainerBase   `json:"-"`
	DeclarationBase `json:"-"`
	ID              *Identifier                `json:"id,omitempty"`
	TypeParameters  *TypeParametersDeclaration `json:"type_parameters,omitempty"`
	Params          []*Node                    `json:"params,omitempty"`
	Body            *BlockStatement            `json:"body,omitempty"`
	TypeAnnotation  *TypeAnnotation            `json:"type_annotation,omitempty"`
}

func NewFunctionExpression(id *Identifier, typeParameters *TypeParametersDeclaration, params []*Node, body *BlockStatement, typeAnnotation *TypeAnnotation) *FunctionExpression {
	return &FunctionExpression{
		ID:             id,
		TypeParameters: typeParameters,
		Params:         params,
		Body:           body,
		TypeAnnotation: typeAnnotation,
	}
}

func (functionExpression *FunctionExpression) MarshalJSON() ([]byte, error) {
	type Alias FunctionExpression
	return json.Marshal(
		wrapNode(
			*functionExpression.AsNode(),
			(*Alias)(functionExpression),
		),
	)
}

func (functionExpression *FunctionExpression) NodeType() NodeType {
	return NodeTypeFunctionExpression
}

func (functionExpression *FunctionExpression) ForEachChild(v Visitor) bool {
	return (functionExpression.ID != nil && visit(v, functionExpression.ID.AsNode())) ||
		visitNodes(v, functionExpression.Params) ||
		(functionExpression.Body != nil && visit(v, functionExpression.Body.AsNode())) ||
		(functionExpression.TypeAnnotation != nil && visit(v, functionExpression.TypeAnnotation.AsNode()))
}

func (functionExpression *FunctionExpression) ContainerBaseData() *ContainerBase {
	return &functionExpression.ContainerBase
}

type ModifierList struct {
	ModifierFlags ModifierFlags
}

func NewModifierList(ModifierFlags ModifierFlags) *ModifierList {
	return &ModifierList{ModifierFlags: ModifierFlags}
}

type IndexSignatureDeclaration struct {
	NodeBase
	DeclarationBase `json:"-"`
	ModifiersBase   `json:"-"`
	Index           *Identifier `json:"index,omitempty"`
	Type            *Node       `json:"type,omitempty"`
}

func NewIndexSignatureDeclaration(index *Identifier, typeNode *Node) *IndexSignatureDeclaration {
	return &IndexSignatureDeclaration{
		Index: index,
		Type:  typeNode,
	}
}

func (indexSignatureDeclaration *IndexSignatureDeclaration) MarshalJSON() ([]byte, error) {
	type Alias IndexSignatureDeclaration
	return json.Marshal(
		wrapNode(
			*indexSignatureDeclaration.AsNode(),
			(*Alias)(indexSignatureDeclaration),
		),
	)
}

func (indexSignatureDeclaration *IndexSignatureDeclaration) NodeType() NodeType {
	return NodeTypeIndexSignatureDeclaration
}

func (indexSignatureDeclaration *IndexSignatureDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, indexSignatureDeclaration.Index.AsNode())
}

type RegularExpressionLiteral struct {
	NodeBase
	Text string `json:"text,omitempty"`
}

func NewRegularExpressionLiteral(text string) *RegularExpressionLiteral {
	return &RegularExpressionLiteral{
		Text: text,
	}
}

func (regularExpressionLiteral *RegularExpressionLiteral) MarshalJSON() ([]byte, error) {
	type Alias RegularExpressionLiteral
	return json.Marshal(
		wrapNode(
			*regularExpressionLiteral.AsNode(),
			(*Alias)(regularExpressionLiteral),
		),
	)
}

func (regularExpressionLiteral *RegularExpressionLiteral) NodeType() NodeType {
	return NodeTypeIndexSignatureDeclaration
}

type ArrowFunction struct {
	NodeBase
	ContainerBase   `json:"-"`
	DeclarationBase `json:"-"`
	TypeParameters  *TypeParametersDeclaration `json:"type_parameters,omitempty"`
	Params          []*Node                    `json:"params,omitempty"`
	Body            *BlockStatement            `json:"body,omitempty"`
	TypeAnnotation  *TypeAnnotation            `json:"type_annotation,omitempty"`
}

func NewArrowFunction(typeParameters *TypeParametersDeclaration, params []*Node, body *BlockStatement, typeAnnotation *TypeAnnotation) *ArrowFunction {
	return &ArrowFunction{
		TypeParameters: typeParameters,
		Params:         params,
		Body:           body,
		TypeAnnotation: typeAnnotation,
	}
}

func (arrowFunction *ArrowFunction) MarshalJSON() ([]byte, error) {
	type Alias ArrowFunction
	return json.Marshal(
		wrapNode(
			*arrowFunction.AsNode(),
			(*Alias)(arrowFunction),
		),
	)
}

func (arrowFunction *ArrowFunction) NodeType() NodeType {
	return NodeTypeFunctionExpression
}

func (arrowFunction *ArrowFunction) ForEachChild(v Visitor) bool {
	return visitNodes(v, arrowFunction.Params) ||
		(arrowFunction.Body != nil && visit(v, arrowFunction.Body.AsNode())) ||
		(arrowFunction.TypeAnnotation != nil && visit(v, arrowFunction.TypeAnnotation.AsNode()))
}

func (arrowFunction *ArrowFunction) ContainerBaseData() *ContainerBase {
	return &arrowFunction.ContainerBase
}

type Parameter struct {
	NodeBase
	DeclarationBase `json:"-"`
	Name            *Node           `json:"name,omitempty"`
	TypeAnnotation  *TypeAnnotation `json:"type_annotation,omitempty"`
	Initializer     *Initializer    `json:"initializer,omitempty"`
}

func NewParameter(name *Node, typeAnnotation *TypeAnnotation, initializer *Initializer) *Parameter {
	return &Parameter{
		Name:           name,
		TypeAnnotation: typeAnnotation,
		Initializer:    initializer,
	}
}

func (parameter *Parameter) MarshalJSON() ([]byte, error) {
	type Alias Parameter
	return json.Marshal(
		wrapNode(
			*parameter.AsNode(),
			(*Alias)(parameter),
		),
	)
}

func (parameter *Parameter) NodeType() NodeType {
	return NodeTypeParameter
}

func (parameter *Parameter) ForEachChild(v Visitor) bool {
	return visit(v, parameter.Name) ||
		(parameter.TypeAnnotation != nil && visit(v, parameter.TypeAnnotation.AsNode())) ||
		(parameter.Initializer != nil && visit(v, parameter.Initializer.AsNode()))
}

type ArrayBindingPattern struct {
	NodeBase
	Elements []*Node `json:"elements,omitempty"`
}

func NewArrayBindingPattern(elements []*Node) *ArrayBindingPattern {
	return &ArrayBindingPattern{
		Elements: elements,
	}
}

func (arrayBindingPattern *ArrayBindingPattern) MarshalJSON() ([]byte, error) {
	type Alias ArrayBindingPattern
	return json.Marshal(
		wrapNode(
			*arrayBindingPattern.AsNode(),
			(*Alias)(arrayBindingPattern),
		),
	)
}

func (arrayBindingPattern *ArrayBindingPattern) NodeType() NodeType {
	return NodeTypeArrayBindingPattern
}

func (arrayBindingPattern *ArrayBindingPattern) ForEachChild(v Visitor) bool {
	return visitNodes(v, arrayBindingPattern.Elements)
}

type BindingElement struct {
	NodeBase
	Element        *Node           `json:"element,omitempty"`
	Rest           bool            `json:"rest,omitempty"`
	TypeAnnotation *TypeAnnotation `json:"type_annotation,omitempty"`
	Initializer    *Initializer    `json:"initializer,omitempty"`
}

func NewBindingElement(element *Node, rest bool, typeAnnotation *TypeAnnotation, initializer *Initializer) *BindingElement {
	return &BindingElement{
		Element:        element,
		Rest:           rest,
		TypeAnnotation: typeAnnotation,
		Initializer:    initializer,
	}
}

func (bindingElement *BindingElement) MarshalJSON() ([]byte, error) {
	type Alias BindingElement
	return json.Marshal(
		wrapNode(
			*bindingElement.AsNode(),
			(*Alias)(bindingElement),
		),
	)
}

func (bindingElement *BindingElement) NodeType() NodeType {
	return NodeTypeBindingElement
}

func (bindingElement *BindingElement) ForEachChild(v Visitor) bool {
	return visit(v, bindingElement.Element) ||
		(bindingElement.TypeAnnotation != nil && visit(v, bindingElement.TypeAnnotation.AsNode())) ||
		(bindingElement.Initializer != nil && visit(v, bindingElement.Initializer.AsNode()))
}

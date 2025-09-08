package ast

import "encoding/json"

type BinaryExpressionOperator = string

const (
	PLUS               BinaryExpressionOperator = "+"
	MINUS              BinaryExpressionOperator = "-"
	SLASH              BinaryExpressionOperator = "/"
	PERCENT            BinaryExpressionOperator = "%"
	STAR               BinaryExpressionOperator = "*"
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
}

type NodeBase struct {
	Node `json:"-"`
}

func NewNode[T NodeData](nodeData T, location Location) T {
	node := nodeData.AsNode()
	node.Type = nodeData.NodeType()
	node.Location = location
	node.Data = nodeData

	node.Data.ForEachChild(func(n *Node) bool {
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

func (node *Node) AsIfStatement() *IfStatement         { return node.Data.(*IfStatement) }
func (node *Node) AsForStatement() *ForStatement       { return node.Data.(*ForStatement) }
func (node *Node) AsBlockStatement() *BlockStatement   { return node.Data.(*BlockStatement) }
func (node *Node) AsSourceFile() *SourceFile           { return node.Data.(*SourceFile) }
func (node *Node) AsReturnStatement() *ReturnStatement { return node.Data.(*ReturnStatement) }
func (node *Node) AsExpressionStatement() *ExpressionStatement {
	return node.Data.(*ExpressionStatement)
}

func (node *Node) AsTInterfaceDeclaration() *TInterfaceDeclaration {
	return node.Data.(*TInterfaceDeclaration)
}

func (node *Node) AsTInterfaceBody() *TInterfaceBody {
	return node.Data.(*TInterfaceBody)
}

func (node *Node) AsTPropertySignature() *TPropertySignature {
	return node.Data.(*TPropertySignature)
}

func (node *Node) AsTTypeReference() *TTypeReference {
	return node.Data.(*TTypeReference)
}

func (node *Node) ForEachChild(v Visitor) bool     { return node.Data.ForEachChild(v) }
func (node *NodeBase) ForEachChild(v Visitor) bool { return false }

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
	DeclarationBase
	Identifier  *Identifier  `json:"identifier,omitempty"`
	Initializer *Initializer `json:"initializer,omitempty"`
	Declare     bool         `json:"declare,omitempty"`
}

func NewVariableDeclaration(identifier *Identifier, initializer *Initializer, declare bool) *VariableDeclaration {
	return &VariableDeclaration{Identifier: identifier, Initializer: initializer, Declare: declare}
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
	return visit(v, variableDeclaration.Identifier.AsNode()) || visit(v, variableDeclaration.Initializer.AsNode())
}

type TTypeAnnotation struct {
	NodeBase
	TypeAnnotation *Node `json:"type_annotation,omitempty"`
}

func NewTTypeAnnotation(typeAnnotation *Node) *TTypeAnnotation {
	return &TTypeAnnotation{TypeAnnotation: typeAnnotation}
}

func (tTypeAnnotation *TTypeAnnotation) MarshalJSON() ([]byte, error) {
	type Alias TTypeAnnotation
	return json.Marshal(
		wrapNode(
			*tTypeAnnotation.AsNode(),
			(*Alias)(tTypeAnnotation),
		),
	)
}

func (tTypeAnnotation *TTypeAnnotation) NodeType() NodeType {
	return NodeTypeVariableDeclaration
}

func (tTypeAnnotation *TTypeAnnotation) ForEachChild(v Visitor) bool {
	return visit(v, tTypeAnnotation.TypeAnnotation)
}

type TStringKeyword struct {
	NodeBase
}

func NewTStringKeyword() *TStringKeyword {
	return &TStringKeyword{}
}

func (tStringKeyword *TStringKeyword) MarshalJSON() ([]byte, error) {
	type Alias TStringKeyword
	return json.Marshal(
		wrapNode(
			*tStringKeyword.AsNode(),
			(*Alias)(tStringKeyword),
		),
	)
}

func (tStringKeyword *TStringKeyword) NodeType() NodeType {
	return NodeTypeTStringKeyword
}

type TNumberKeyword struct {
	NodeBase
}

func NewTNumberKeyword() *TNumberKeyword {
	return &TNumberKeyword{}
}

func (tNumberKeyword *TNumberKeyword) MarshalJSON() ([]byte, error) {
	type Alias TNumberKeyword
	return json.Marshal(
		wrapNode(
			*tNumberKeyword.AsNode(),
			(*Alias)(tNumberKeyword),
		),
	)
}

func (tNumberKeyword *TNumberKeyword) NodeType() NodeType {
	return NodeTypeTNumberKeyword
}

type TBooleanKeyword struct {
	NodeBase
}

func NewTBooleanKeyword() *TBooleanKeyword {
	return &TBooleanKeyword{}
}

func (tBooleanKeyword *TBooleanKeyword) MarshalJSON() ([]byte, error) {
	type Alias TBooleanKeyword
	return json.Marshal(
		wrapNode(
			*tBooleanKeyword.AsNode(),
			(*Alias)(tBooleanKeyword),
		),
	)
}

func (tBooleanKeyword *TBooleanKeyword) NodeType() NodeType {
	return NodeTypeTBooleanKeyword
}

type TNullKeyword struct {
	NodeBase
}

func NewTNullKeyword() *TNullKeyword {
	return &TNullKeyword{}
}

func (tNullKeyword *TNullKeyword) MarshalJSON() ([]byte, error) {
	type Alias TNullKeyword
	return json.Marshal(
		wrapNode(
			*tNullKeyword.AsNode(),
			(*Alias)(tNullKeyword),
		),
	)
}

func (tNullKeyword *TNullKeyword) NodeType() NodeType {
	return NodeTypeTNullKeyword
}

type TAnyKeyword struct {
	NodeBase
}

func NewTAnyKeyword() *TAnyKeyword {
	return &TAnyKeyword{}
}

func (tAnyKeyword *TAnyKeyword) MarshalJSON() ([]byte, error) {
	type Alias TAnyKeyword
	return json.Marshal(
		wrapNode(
			*tAnyKeyword.AsNode(),
			(*Alias)(tAnyKeyword),
		),
	)
}

func (tAnyKeyword *TAnyKeyword) NodeType() NodeType {
	return NodeTypeTAnyKeyword
}

type Identifier struct {
	NodeBase
	Name           string           `json:"name,omitempty"`
	OriginalName   *string          `json:"original_name,omitempty"`
	TypeAnnotation *TTypeAnnotation `json:"type_annotation,omitempty"`
}

func NewIdentifier(name string, typeAnnotation *TTypeAnnotation) *Identifier {
	return &Identifier{Name: name, TypeAnnotation: typeAnnotation, OriginalName: nil}
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
	return visit(v, identifier.TypeAnnotation.AsNode())
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
	Scope         *Scope  `json:"-"`
}

func NewBlockStatement(body []*Node) *BlockStatement {
	return &BlockStatement{Body: body, Scope: &Scope{}}
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
	ID              *Identifier      `json:"id,omitempty"`
	Params          []*Node          `json:"params,omitempty"`
	Body            *BlockStatement  `json:"body,omitempty"`
	TTypeAnnotation *TTypeAnnotation `json:"type_annotation,omitempty"`
}

func NewFunctionDeclaration(id *Identifier, params []*Node, body *BlockStatement, tTypeAnnotation *TTypeAnnotation) *FunctionDeclaration {
	return &FunctionDeclaration{
		ID:              id,
		Params:          params,
		Body:            body,
		TTypeAnnotation: tTypeAnnotation,
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
		visit(v, functionDeclaration.Body.AsNode()) ||
		visit(v, functionDeclaration.TTypeAnnotation.AsNode())
}

type CallExpression struct {
	NodeBase
	Callee *Node   `json:"callee,omitempty"`
	Args   []*Node `json:"args,omitempty"`
}

func NewCallExpression(callee *Node, args []*Node) *CallExpression {
	return &CallExpression{Callee: callee, Args: args}
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
	return visit(v, callExpression.Callee) || visitNodes(v, callExpression.Args)
}

type MemberExpression struct {
	NodeBase
	Object   *Node `json:"object,omitempty"`
	Property *Node `json:"property,omitempty"`
}

func NewMemberExpression(object *Node, property *Node) *MemberExpression {
	return &MemberExpression{Object: object, Property: property}
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

	ExternalModuleIndicator *Node `json:"-"`
}

func NewSourceFile(body []*Node, directives []*Directive) *SourceFile {
	return &SourceFile{Name: "", Body: body, Directives: directives}
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

type TInterfaceDeclaration struct {
	NodeBase
	Id   *Identifier     `json:"id,omitempty"`
	Body *TInterfaceBody `json:"body,omitempty"`
}

func NewTInterfaceDeclaration(id *Identifier, body *TInterfaceBody) *TInterfaceDeclaration {
	return &TInterfaceDeclaration{
		Id:   id,
		Body: body,
	}
}

func (tInterfaceDeclaration *TInterfaceDeclaration) MarshalJSON() ([]byte, error) {
	type Alias TInterfaceDeclaration
	return json.Marshal(
		wrapNode(
			*tInterfaceDeclaration.AsNode(),
			(*Alias)(tInterfaceDeclaration),
		),
	)
}

func (tInterfaceDeclaration *TInterfaceDeclaration) NodeType() NodeType {
	return NodeTypeTInterfaceDeclaration
}

func (tInterfaceDeclaration *TInterfaceDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, tInterfaceDeclaration.Id.AsNode()) || visit(v, tInterfaceDeclaration.Body.AsNode())
}

type TInterfaceBody struct {
	NodeBase
	Body []*Node `json:"body,omitempty"`
}

func NewTInterfaceBody(body []*Node) *TInterfaceBody {
	return &TInterfaceBody{Body: body}
}

func (tInterfaceBody *TInterfaceBody) MarshalJSON() ([]byte, error) {
	type Alias TInterfaceBody
	return json.Marshal(
		wrapNode(
			*tInterfaceBody.AsNode(),
			(*Alias)(tInterfaceBody),
		),
	)
}

func (tInterfaceBody *TInterfaceBody) NodeType() NodeType {
	return NodeTypeTInterfaceBody
}

func (tInterfaceBody *TInterfaceBody) ForEachChild(v Visitor) bool {
	return visitNodes(v, tInterfaceBody.Body)
}

type TPropertySignature struct {
	NodeBase
	Key            *Node            `json:"key,omitempty"`
	TypeAnnotation *TTypeAnnotation `json:"type_annotation,omitempty"`
}

func NewTPropertySignature(key *Node, typeAnnotation *TTypeAnnotation) *TPropertySignature {
	return &TPropertySignature{
		Key:            key,
		TypeAnnotation: typeAnnotation,
	}
}

func (tPropertySignature *TPropertySignature) MarshalJSON() ([]byte, error) {
	type Alias TPropertySignature
	return json.Marshal(
		wrapNode(
			*tPropertySignature.AsNode(),
			(*Alias)(tPropertySignature),
		),
	)
}

func (tPropertySignature *TPropertySignature) NodeType() NodeType {
	return NodeTypeTPropertySignature
}

func (tPropertySignature *TPropertySignature) ForEachChild(v Visitor) bool {
	return visit(v, tPropertySignature.Key) || visit(v, tPropertySignature.TypeAnnotation.AsNode())
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

type TFunctionType struct {
	NodeBase
	Params         []*Node          `json:"params,omitempty"`
	TypeAnnotation *TTypeAnnotation `json:"type_annotation,omitempty"`
}

func NewTFunctionType(params []*Node, typeAnnotation *TTypeAnnotation) *TFunctionType {
	return &TFunctionType{
		Params:         params,
		TypeAnnotation: typeAnnotation,
	}
}

func (tFunctionType *TFunctionType) MarshalJSON() ([]byte, error) {
	type Alias TFunctionType
	return json.Marshal(
		wrapNode(
			*tFunctionType.AsNode(),
			(*Alias)(tFunctionType),
		),
	)
}

func (tFunctionType *TFunctionType) NodeType() NodeType {
	return NodeTypeTFunctionType
}

func (tFunctionType *TFunctionType) ForEachChild(v Visitor) bool {
	return visitNodes(v, tFunctionType.Params) || visit(v, tFunctionType.TypeAnnotation.AsNode())
}

type TTypeAliasDeclaration struct {
	NodeBase
	Id             *Identifier      `json:"id,omitempty"`
	TypeAnnotation *TTypeAnnotation `json:"type_annotation,omitempty"`
}

func NewTTypeAliasDeclaration(id *Identifier, typeAnnotation *TTypeAnnotation) *TTypeAliasDeclaration {
	return &TTypeAliasDeclaration{
		Id:             id,
		TypeAnnotation: typeAnnotation,
	}
}

func (tTypeAliasDeclaration *TTypeAliasDeclaration) MarshalJSON() ([]byte, error) {
	type Alias TTypeAliasDeclaration
	return json.Marshal(
		wrapNode(
			*tTypeAliasDeclaration.AsNode(),
			(*Alias)(tTypeAliasDeclaration),
		),
	)
}

func (tTypeAliasDeclaration *TTypeAliasDeclaration) NodeType() NodeType {
	return NodeTypeTFunctionType
}

func (tTypeAliasDeclaration *TTypeAliasDeclaration) ForEachChild(v Visitor) bool {
	return visit(v, tTypeAliasDeclaration.Id.AsNode()) || visit(v, tTypeAliasDeclaration.TypeAnnotation.AsNode())
}

type TTypeLiteral struct {
	NodeBase
	Members []*Node `json:"members,omitempty"`
}

func NewTTypeLiteral(members []*Node) *TTypeLiteral {
	return &TTypeLiteral{Members: members}
}

func (tTypeLiteral *TTypeLiteral) MarshalJSON() ([]byte, error) {
	type Alias TTypeLiteral
	return json.Marshal(
		wrapNode(
			*tTypeLiteral.AsNode(),
			(*Alias)(tTypeLiteral),
		),
	)
}

func (tTypeLiteral *TTypeLiteral) NodeType() NodeType {
	return NodeTypeTFunctionType
}

func (tTypeLiteral *TTypeLiteral) ForEachChild(v Visitor) bool {
	return visitNodes(v, tTypeLiteral.Members)
}

type TTypeReference struct {
	NodeBase
	TypeName *Identifier `json:"type_name,omitempty"`
}

func NewTTypeReference(TypeName *Identifier) *TTypeReference {
	return &TTypeReference{TypeName: TypeName}
}

func (tTypeReference *TTypeReference) MarshalJSON() ([]byte, error) {
	type Alias TTypeReference
	return json.Marshal(
		wrapNode(
			*tTypeReference.AsNode(),
			(*Alias)(tTypeReference),
		),
	)
}

func (tTypeReference *TTypeReference) NodeType() NodeType {
	return NodeTypeTTypeReference
}

func (tTypeReference *TTypeReference) ForEachChild(v Visitor) bool {
	return visit(v, tTypeReference.TypeName.AsNode())
}

type TArrayType struct {
	NodeBase
	ElementType *Node `json:"element_type,omitempty"`
}

func NewTArrayType(elementType *Node) *TArrayType {
	return &TArrayType{ElementType: elementType}
}

func (tArrayType *TArrayType) MarshalJSON() ([]byte, error) {
	type Alias TArrayType
	return json.Marshal(
		wrapNode(
			*tArrayType.AsNode(),
			(*Alias)(tArrayType),
		),
	)
}

func (tArrayType *TArrayType) NodeType() NodeType {
	return NodeTypeTArrayType
}

func (tArrayType *TArrayType) ForEachChild(v Visitor) bool {
	return visit(v, tArrayType.AsNode())
}

type RestElement struct {
	NodeBase
	Argument       *Identifier      `json:"argument,omitempty"`
	TypeAnnotation *TTypeAnnotation `json:"type_annotation,omitempty"`
}

func NewRestElement(argument *Identifier, typeAnnotation *TTypeAnnotation) *RestElement {
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
	return visit(v, restElement.Argument.AsNode()) || visit(v, restElement.TypeAnnotation.AsNode())
}

type ForStatement struct {
	NodeBase
	Init   *Node `json:"init,omitempty"`
	Test   *Node `json:"test,omitempty"`
	Update *Node `json:"update,omitempty"`
	Body   *Node `json:"body,omitempty"`
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

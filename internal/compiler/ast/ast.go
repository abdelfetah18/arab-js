package ast

type JavaScriptNode interface{}
type LVal interface{}
type FunctionParameter = Identifier

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

type Node struct {
	Type NodeType
	Data NodeData
}

type NodeData interface{}

func (node *Node) AsVariableDeclaration() *VariableDeclaration {
	return node.Data.(*VariableDeclaration)
}

func (node *Node) AsDecimalLiteral() *DecimalLiteral { return node.Data.(*DecimalLiteral) }
func (node *Node) AsStringLiteral() *StringLiteral   { return node.Data.(*StringLiteral) }
func (node *Node) AsBooleanLiteral() *BooleanLiteral { return node.Data.(*BooleanLiteral) }
func (node *Node) AsNullLiteral() *NullLiteral       { return node.Data.(*NullLiteral) }
func (node *Node) AsCallExpression() *CallExpression { return node.Data.(*CallExpression) }
func (node *Node) AsIdentifier() *Identifier         { return node.Data.(*Identifier) }

func (node *Node) AsIfStatement() *IfStatement       { return node.Data.(*IfStatement) }
func (node *Node) AsBlockStatement() *BlockStatement { return node.Data.(*BlockStatement) }
func (node *Node) AsProgram() *Program               { return node.Data.(*Program) }
func (node *Node) AsExpressionStatement() *ExpressionStatement {
	return node.Data.(*ExpressionStatement)
}

type ExpressionStatement struct {
	Expression *Node
}

func NewExpressionStatement(expression *Node) *ExpressionStatement {
	return &ExpressionStatement{Expression: expression}
}

func (expressionStatement *ExpressionStatement) ToNode() *Node {
	return &Node{
		Type: NodeTypeExpressionStatement,
		Data: expressionStatement,
	}
}

type VariableDeclaration struct {
	Identifier     *Identifier
	Initializer    *Initializer
	TypeAnnotation *TTypeAnnotation
}

func NewVariableDeclaration(identifier *Identifier, initializer *Initializer, typeAnnotation *TTypeAnnotation) *VariableDeclaration {
	return &VariableDeclaration{Identifier: identifier, Initializer: initializer, TypeAnnotation: typeAnnotation}
}

func (variableDeclaration *VariableDeclaration) ToNode() *Node {
	return &Node{
		Type: NodeTypeVariableDeclaration,
		Data: variableDeclaration,
	}
}

type TTypeAnnotation struct {
	TypeAnnotation *Node // TStringKeyword | TNumberKeyword | TBooleanKeyword | TNullKeyword
}

func NewTTypeAnnotation(typeAnnotation *Node) *TTypeAnnotation {
	return &TTypeAnnotation{TypeAnnotation: typeAnnotation}
}

func (tTypeAnnotation *TTypeAnnotation) ToNode() *Node {
	return &Node{
		Type: NodeTypeVariableDeclaration,
		Data: tTypeAnnotation,
	}
}

type TStringKeyword struct{}

func NewTStringKeyword() *TStringKeyword {
	return &TStringKeyword{}
}

func (tStringKeyword TStringKeyword) ToNode() *Node {
	return &Node{
		Type: NodeTypeTStringKeyword,
		Data: tStringKeyword,
	}
}

type TNumberKeyword struct{}

func NewTNumberKeyword() *TNumberKeyword {
	return &TNumberKeyword{}
}

func (tNumberKeyword TNumberKeyword) ToNode() *Node {
	return &Node{
		Type: NodeTypeTNumberKeyword,
		Data: tNumberKeyword,
	}
}

type TBooleanKeyword struct{}

func NewTBooleanKeyword() *TBooleanKeyword {
	return &TBooleanKeyword{}
}

func (tBooleanKeyword TBooleanKeyword) ToNode() *Node {
	return &Node{
		Type: NodeTypeTBooleanKeyword,
		Data: tBooleanKeyword,
	}
}

type TNullKeyword struct{}

func NewTNullKeyword() *TNullKeyword {
	return &TNullKeyword{}
}

func (tNullKeyword TNullKeyword) ToNode() *Node {
	return &Node{
		Type: NodeTypeTNullKeyword,
		Data: tNullKeyword,
	}
}

type Identifier struct {
	Name string
}

func NewIdentifier(name string) *Identifier {
	return &Identifier{Name: name}
}

func (identifier *Identifier) ToNode() *Node {
	return &Node{
		Type: NodeTypeIdentifier,
		Data: identifier,
	}
}

type Initializer struct {
	Expression *Node
}

func NewInitializer(expression *Node) *Initializer {
	return &Initializer{Expression: expression}
}

func (initializer *Initializer) ToNode() *Node {
	return &Node{
		Type: NodeTypeInitializer,
		Data: initializer,
	}
}

type StringLiteral struct {
	Value string
}

func NewStringLiteral(value string) *StringLiteral {
	return &StringLiteral{Value: value}
}

func (stringLiteral *StringLiteral) ToNode() *Node {
	return &Node{
		Type: NodeTypeStringLiteral,
		Data: stringLiteral,
	}
}

type NullLiteral struct{}

func NewNullLiteral() *NullLiteral {
	return &NullLiteral{}
}

func (nullLiteral *NullLiteral) ToNode() *Node {
	return &Node{
		Type: NodeTypeNullLiteral,
		Data: nullLiteral,
	}
}

type BooleanLiteral struct {
	Value bool
}

func NewBooleanLiteral(value bool) *BooleanLiteral {
	return &BooleanLiteral{Value: value}
}

func (booleanLiteral *BooleanLiteral) ToNode() *Node {
	return &Node{
		Type: NodeTypeBooleanLiteral,
		Data: booleanLiteral,
	}
}

type DecimalLiteral struct {
	Value string
}

func NewDecimalLiteral(value string) *DecimalLiteral {
	return &DecimalLiteral{Value: value}
}

func (decimalLiteral *DecimalLiteral) ToNode() *Node {
	return &Node{
		Type: NodeTypeDecimalLiteral,
		Data: decimalLiteral,
	}
}

type IfStatement struct {
	TestExpression      *Node
	ConsequentStatement *Node
	AlternateStatement  *Node
}

func NewIfStatement(test *Node, consequent *Node, alternate *Node) *IfStatement {
	return &IfStatement{TestExpression: test, ConsequentStatement: consequent, AlternateStatement: alternate}
}

func (ifStatement *IfStatement) ToNode() *Node {
	return &Node{
		Type: NodeTypeIfStatement,
		Data: ifStatement,
	}
}

type BlockStatement struct {
	Body []*Node
}

func NewBlockStatement(body []*Node) *BlockStatement {
	return &BlockStatement{Body: body}
}

func (blockStatement *BlockStatement) ToNode() *Node {
	return &Node{
		Type: NodeTypeBlockStatement,
		Data: blockStatement,
	}
}

type AssignmentExpression struct {
	Operator string
	Left     LVal
	Right    *Node
}

func NewAssignmentExpression(operator string, left LVal, right *Node) *AssignmentExpression {
	return &AssignmentExpression{Operator: operator, Left: left, Right: right}
}

func (assignmentExpression *AssignmentExpression) ToNode() *Node {
	return &Node{
		Type: NodeTypeAssignmentExpression,
		Data: assignmentExpression,
	}
}

type FunctionDeclaration struct {
	ID     *Identifier
	Params []*Identifier
	Body   *BlockStatement
}

func NewFunctionDeclaration(id *Identifier, params []*Identifier, body *BlockStatement) *FunctionDeclaration {
	return &FunctionDeclaration{ID: id, Params: params, Body: body}
}

func (functionDeclaration *FunctionDeclaration) ToNode() *Node {
	return &Node{
		Type: NodeTypeFunctionDeclaration,
		Data: functionDeclaration,
	}
}

type CallExpression struct {
	Callee *Node
	Args   []*Node
}

func NewCallExpression(callee *Node, args []*Node) *CallExpression {
	return &CallExpression{Callee: callee, Args: args}
}

func (callExpression *CallExpression) ToNode() *Node {
	return &Node{
		Type: NodeTypeCallExpression,
		Data: callExpression,
	}
}

type MemberExpression struct {
	Object   *Node
	Property *Node
}

func NewMemberExpression(object *Node, property *Node) *MemberExpression {
	return &MemberExpression{Object: object, Property: property}
}

func (memberExpression *MemberExpression) ToNode() *Node {
	return &Node{
		Type: NodeTypeMemberExpression,
		Data: memberExpression,
	}
}

type ImportSpecifier struct {
	Local    *Identifier
	Imported *Node
}

func NewImportSpecifier(local *Identifier, imported *Node) *ImportSpecifier {
	return &ImportSpecifier{Local: local, Imported: imported}
}

func (importSpecifier *ImportSpecifier) ToNode() *Node {
	return &Node{
		Type: NodeTypeImportSpecifier,
		Data: importSpecifier,
	}
}

type ImportDefaultSpecifier struct {
	Local *Identifier
}

func NewImportDefaultSpecifier(local *Identifier) *ImportDefaultSpecifier {
	return &ImportDefaultSpecifier{Local: local}
}

func (importDefaultSpecifier *ImportDefaultSpecifier) ToNode() *Node {
	return &Node{
		Type: NodeTypeImportDefaultSpecifier,
		Data: importDefaultSpecifier,
	}
}

type ImportNamespaceSpecifier struct {
	Local *Identifier
}

func NewImportNamespaceSpecifier(local *Identifier) *ImportNamespaceSpecifier {
	return &ImportNamespaceSpecifier{Local: local}
}

func (importNamespaceSpecifier *ImportNamespaceSpecifier) ToNode() *Node {
	return &Node{
		Type: NodeTypeImportNamespaceSpecifier,
		Data: importNamespaceSpecifier,
	}
}

type ImportSpecifierInterface interface{}

type ImportDeclaration struct {
	Specifiers []ImportSpecifierInterface
	Source     *StringLiteral
}

func NewImportDeclaration(specifiers []ImportSpecifierInterface, source *StringLiteral) *ImportDeclaration {
	return &ImportDeclaration{Specifiers: specifiers, Source: source}
}

func (importDeclaration *ImportDeclaration) ToNode() *Node {
	return &Node{
		Type: NodeTypeImportDeclaration,
		Data: importDeclaration,
	}
}

type ExportNamedDeclaration struct {
	Declaration *Node
	Specifiers  []*ExportSpecifier
	Source      *StringLiteral
}

func NewExportNamedDeclaration(declaration *Node, specifiers []*ExportSpecifier, source *StringLiteral) *ExportNamedDeclaration {
	return &ExportNamedDeclaration{Declaration: declaration, Specifiers: specifiers, Source: source}
}

func (exportNamedDeclaration *ExportNamedDeclaration) ToNode() *Node {
	return &Node{
		Type: NodeTypeExportNamedDeclaration,
		Data: exportNamedDeclaration,
	}
}

type ExportDefaultDeclaration struct {
	Declaration *Node
}

func NewExportDefaultDeclaration(declaration *Node) *ExportDefaultDeclaration {
	return &ExportDefaultDeclaration{Declaration: declaration}
}

func (exportDefaultDeclaration *ExportDefaultDeclaration) ToNode() *Node {
	return &Node{
		Type: NodeTypeExportDefaultDeclaration,
		Data: exportDefaultDeclaration,
	}
}

type ExportSpecifier struct {
	Local    *Identifier
	Exported *Node
}

func NewExportSpecifier(local *Identifier, exported *Node) *ExportSpecifier {
	return &ExportSpecifier{Local: local, Exported: exported}
}

func (exportSpecifier *ExportSpecifier) ToNode() *Node {
	return &Node{
		Type: NodeTypeExportSpecifier,
		Data: exportSpecifier,
	}
}

type BinaryExpression struct {
	Operator BinaryExpressionOperator
	Left     *Node
	Right    *Node
}

func NewBinaryExpression(operator BinaryExpressionOperator, left *Node, right *Node) *BinaryExpression {
	return &BinaryExpression{Operator: operator, Left: left, Right: right}
}

func (binaryExpression *BinaryExpression) ToNode() *Node {
	return &Node{
		Type: NodeTypeBinaryExpression,
		Data: binaryExpression,
	}
}

type SpreadElement struct {
	Argument *Node
}

func NewSpreadElement(argument *Node) *SpreadElement {
	return &SpreadElement{Argument: argument}
}

func (spreadElement *SpreadElement) ToNode() *Node {
	return &Node{
		Type: NodeTypeSpreadElement,
		Data: spreadElement,
	}
}

type ArrayExpression struct {
	Elements []*Node
}

func NewArrayExpression(elements []*Node) *ArrayExpression {
	return &ArrayExpression{Elements: elements}
}

func (arrayExpression *ArrayExpression) ToNode() *Node {
	return &Node{
		Type: NodeTypeArrayExpression,
		Data: arrayExpression,
	}
}

type ObjectProperty struct {
	Key   *Node
	Value *Node
}

func NewObjectProperty(key *Node, value *Node) *ObjectProperty {
	return &ObjectProperty{Key: key, Value: value}
}

func (objectProperty *ObjectProperty) ToNode() *Node {
	return &Node{
		Type: NodeTypeObjectProperty,
		Data: objectProperty,
	}
}

type ObjectMethod struct {
	Kind   string
	Key    *Node
	Params []*Identifier
	Body   *BlockStatement
}

func NewObjectMethod(kind string, key *Node, params []*Identifier, body *BlockStatement) *ObjectMethod {
	return &ObjectMethod{Kind: kind, Key: key, Params: params, Body: body}
}

func (objectMethod *ObjectMethod) ToNode() *Node {
	return &Node{
		Type: NodeTypeObjectMethod,
		Data: objectMethod,
	}
}

type ObjectPropertyInterface interface{}

type ObjectExpression struct {
	Properties []ObjectPropertyInterface
}

func NewObjectExpression(properties []ObjectPropertyInterface) *ObjectExpression {
	return &ObjectExpression{Properties: properties}
}

func (objectExpression *ObjectExpression) ToNode() *Node {
	return &Node{
		Type: NodeTypeObjectExpression,
		Data: objectExpression,
	}
}

type DirectiveLiteral struct {
	Value string
}

func NewDirectiveLiteral(value string) *DirectiveLiteral {
	return &DirectiveLiteral{Value: value}
}

func (directiveLiteral *DirectiveLiteral) ToNode() *Node {
	return &Node{
		Type: NodeTypeDirectiveLiteral,
		Data: directiveLiteral,
	}
}

type Directive struct {
	Value *DirectiveLiteral
}

func NewDirective(value *DirectiveLiteral) *Directive {
	return &Directive{Value: value}
}

func (directive *Directive) ToNode() *Node {
	return &Node{
		Type: NodeTypeDirective,
		Data: directive,
	}
}

type Program struct {
	Body       []*Node
	Directives []*Directive
}

func NewProgram(body []*Node, directives []*Directive) *Program {
	return &Program{Body: body, Directives: directives}
}

func (program *Program) ToNode() *Node {
	return &Node{
		Type: NodeTypeProgram,
		Data: program,
	}
}

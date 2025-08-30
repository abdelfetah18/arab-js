package checker

import (
	"arab_js/internal/compiler/ast"
	"fmt"
)

type Checker struct {
	Program     *ast.Program
	SymbolTable *SymbolTable
	Errors      []string
}

func NewChecker(program *ast.Program) *Checker {
	return &Checker{
		Program:     program,
		SymbolTable: BuildSymbolTable(program),
		Errors:      []string{},
	}
}

func (c *Checker) error(msg string) {
	c.Errors = append(c.Errors, msg)
}

func (c *Checker) errorf(format string, a ...any) {
	c.error(fmt.Sprintf(format, a...))
}

func (c *Checker) Check() {
	for _, node := range c.Program.Body {
		c.checkStatement(node)
	}
}

func (c *Checker) checkStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeVariableDeclaration:
		c.checkVariableDeclaration(node.AsVariableDeclaration())
	}
}

func (c *Checker) checkVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Identifier.TypeAnnotation == nil {
		// NOTE: Identifier is any so we skip
		return
	}

	identifierType := getTypeFromTypeAnnotationNode(variableDeclaration.Identifier.TypeAnnotation)
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)

	if identifierType.data.Name() != initializerType.data.Name() {
		c.errorf(TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.data.Name(), initializerType.data.Name())
		return
	}
}

func (c *Checker) checkExpression(expression *ast.Node) *Type {
	switch expression.Type {
	case ast.NodeTypeStringLiteral:
		return inferTypeFromNode(expression.AsStringLiteral().ToNode())
	case ast.NodeTypeDecimalLiteral:
		return inferTypeFromNode(expression.AsDecimalLiteral().ToNode())
	case ast.NodeTypeBooleanLiteral:
		return inferTypeFromNode(expression.AsBooleanLiteral().ToNode())
	case ast.NodeTypeNullLiteral:
		return inferTypeFromNode(expression.AsNullLiteral().ToNode())
	}

	return nil
}

type Type struct {
	data TypeData
}

type TypeData interface {
	Name() string
}

func (t *Type) AsStringType() *StringType { return t.data.(*StringType) }

type StringType struct{}

func NewStringType() *StringType    { return &StringType{} }
func (s *StringType) ToType() *Type { return &Type{data: s} }
func (s *StringType) Name() string  { return "string" }

type NumberType struct{}

func NewNumberType() *NumberType    { return &NumberType{} }
func (s *NumberType) ToType() *Type { return &Type{data: s} }
func (s *NumberType) Name() string  { return "number" }

type BooleanType struct{}

func NewBooleanType() *BooleanType   { return &BooleanType{} }
func (s *BooleanType) ToType() *Type { return &Type{data: s} }
func (s *BooleanType) Name() string  { return "boolean" }

type NullType struct{}

func NewNullType() *NullType      { return &NullType{} }
func (s *NullType) ToType() *Type { return &Type{data: s} }
func (s *NullType) Name() string  { return "null" }

type Symbol struct {
	Name string
	Type *Type
}

type Scope struct {
	Variables map[string]*Symbol
	Parent    *Scope
}

func (s *Scope) AddVariable(name string, _type *Type) error {
	if s.Variables == nil {
		s.Variables = make(map[string]*Symbol)
	}

	if _, exists := s.Variables[name]; exists {
		return fmt.Errorf("variable '%s' already declared in this scope", name)
	}

	s.Variables[name] = &Symbol{
		Name: name,
		Type: _type,
	}

	return nil
}

type SymbolTable struct {
	Current *Scope
}

func BuildSymbolTable(program *ast.Program) *SymbolTable {
	currentScope := &Scope{
		Variables: map[string]*Symbol{},
		Parent:    nil,
	}

	symbolTable := &SymbolTable{
		Current: currentScope,
	}

	nodeVisitor := ast.NewNodeVisitor(func(node *ast.Node) *ast.Node {
		if node.Type == ast.NodeTypeBlockStatement {
			newScope := &Scope{
				Variables: map[string]*Symbol{},
				Parent:    currentScope,
			}

			currentScope = newScope
		}

		if node.Type == ast.NodeTypeIdentifier {
			identifier := node.AsIdentifier()
			if identifier.TypeAnnotation != nil {
				currentScope.AddVariable(
					identifier.Name,
					getTypeFromTypeAnnotationNode(identifier.TypeAnnotation),
				)
			}
		}

		return nil
	})

	nodeVisitor.VisitNode(program.ToNode())

	return symbolTable
}

func getTypeFromTypeAnnotationNode(node *ast.TTypeAnnotation) *Type {
	return getTypeFromTypeNode(node.TypeAnnotation)
}

func getTypeFromTypeNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeTStringKeyword:
		return NewStringType().ToType()
	case ast.NodeTypeTBooleanKeyword:
		return NewBooleanType().ToType()
	case ast.NodeTypeTNumberKeyword:
		return NewNumberType().ToType()
	case ast.NodeTypeTNullKeyword:
		return NewNullType().ToType()
	}

	return nil
}

func inferTypeFromNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeStringLiteral:
		return NewStringType().ToType()
	case ast.NodeTypeDecimalLiteral:
		return NewNumberType().ToType()
	case ast.NodeTypeBooleanLiteral:
		return NewBooleanType().ToType()
	case ast.NodeTypeNullLiteral:
		return NewNullType().ToType()
	}

	return nil
}

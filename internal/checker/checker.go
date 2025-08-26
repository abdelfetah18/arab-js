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
	if variableDeclaration.Identifier.TypeAnnotation.TypeAnnotation.Type == ast.NodeTypeTNumberKeyword {
		if variableDeclaration.Initializer.Expression.Type != ast.NodeTypeDecimalLiteral {
			c.error("Type 'string' is not assignable to type 'number'")
			return
		}
	}
}

type Type int16

const (
	TypeUnknown Type = iota
	TypeString
	TypeNumber
	TypeBoolean
	TypeNull
)

type Symbol struct {
	Name string
	Type Type
}

type Scope struct {
	Variables map[string]*Symbol
	Parent    *Scope
}

func (s *Scope) AddVariable(name string, variableType Type) error {
	if s.Variables == nil {
		s.Variables = make(map[string]*Symbol)
	}

	if _, exists := s.Variables[name]; exists {
		return fmt.Errorf("variable '%s' already declared in this scope", name)
	}

	s.Variables[name] = &Symbol{
		Name: name,
		Type: variableType,
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

		if node.Type == ast.NodeTypeVariableDeclaration {
			variableDeclaration := node.AsVariableDeclaration()

			typeAnnotation := variableDeclaration.Identifier.TypeAnnotation.TypeAnnotation
			if typeAnnotation.Type == ast.NodeTypeTNumberKeyword {
				currentScope.AddVariable(variableDeclaration.Identifier.Name, TypeNumber)
			}
		}

		return nil
	})

	nodeVisitor.VisitNode(program.ToNode())

	return symbolTable
}

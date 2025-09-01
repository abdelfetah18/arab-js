package ast

import (
	"fmt"
)

type Symbol struct {
	Name         string
	OriginalName *string
	Type         *Type
}

type Scope struct {
	Locals map[string]*Symbol
	Parent *Scope
}

func (s *Scope) AddVariable(name string, originalName *string, _type *Type) error {
	if s.Locals == nil {
		s.Locals = make(map[string]*Symbol)
	}

	if _, exists := s.Locals[name]; exists {
		return fmt.Errorf("variable '%s' already declared in this scope", name)
	}

	s.Locals[name] = &Symbol{
		Name:         name,
		Type:         _type,
		OriginalName: originalName,
	}

	return nil
}

func (s *Scope) GetVariableSymbol(name string) *Symbol {
	if s.Locals == nil {
		return nil
	}

	if symbol, exists := s.Locals[name]; exists {
		return symbol
	}

	return nil
}

type SymbolTable map[string]*Symbol

type Type struct {
	Data TypeData
}

type TypeData interface {
	Name() string
}

func (t *Type) AsStringType() *StringType { return t.Data.(*StringType) }

type StringType struct{}

func NewStringType() *StringType    { return &StringType{} }
func (s *StringType) ToType() *Type { return &Type{Data: s} }
func (s *StringType) Name() string  { return "string" }

type NumberType struct{}

func NewNumberType() *NumberType    { return &NumberType{} }
func (s *NumberType) ToType() *Type { return &Type{Data: s} }
func (s *NumberType) Name() string  { return "number" }

type BooleanType struct{}

func NewBooleanType() *BooleanType   { return &BooleanType{} }
func (s *BooleanType) ToType() *Type { return &Type{Data: s} }
func (s *BooleanType) Name() string  { return "boolean" }

type NullType struct{}

func NewNullType() *NullType      { return &NullType{} }
func (s *NullType) ToType() *Type { return &Type{Data: s} }
func (s *NullType) Name() string  { return "null" }

func GetTypeFromTypeAnnotationNode(node *TTypeAnnotation) *Type {
	return GetTypeFromTypeNode(node.TypeAnnotation)
}

func GetTypeFromTypeNode(node *Node) *Type {
	switch node.Type {
	case NodeTypeTStringKeyword:
		return NewStringType().ToType()
	case NodeTypeTBooleanKeyword:
		return NewBooleanType().ToType()
	case NodeTypeTNumberKeyword:
		return NewNumberType().ToType()
	case NodeTypeTNullKeyword:
		return NewNullType().ToType()
	}

	return nil
}

func InferTypeFromNode(node *Node) *Type {
	switch node.Type {
	case NodeTypeStringLiteral:
		return NewStringType().ToType()
	case NodeTypeDecimalLiteral:
		return NewNumberType().ToType()
	case NodeTypeBooleanLiteral:
		return NewBooleanType().ToType()
	case NodeTypeNullLiteral:
		return NewNullType().ToType()
	}

	return nil
}

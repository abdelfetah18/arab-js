package ast

type Symbol struct {
	Name         string
	OriginalName *string
	Type         *Type
}

type Scope struct {
	Locals map[string]*Symbol
	Parent *Scope
}

func (s *Scope) AddVariable(name string, originalName *string, _type *Type) *Symbol {
	if s.Locals == nil {
		s.Locals = make(map[string]*Symbol)
	}

	if _, exists := s.Locals[name]; exists {
		return nil
	}

	symbol := Symbol{
		Name:         name,
		Type:         _type,
		OriginalName: originalName,
	}
	s.Locals[name] = &symbol
	return &symbol
}

func (s *Scope) GetTypeOfNode(node *Node) *Type {
	switch node.Type {
	case NodeTypeMemberExpression:
		memberExpression := node.AsMemberExpression()
		parent := s.GetTypeOfNode(memberExpression.Object)
		return s.GetTypeOfChildNode(memberExpression.Property, parent.AsObjectType())
	case NodeTypeIdentifier:
		return s.GetVariableSymbol(node.AsIdentifier().Name).Type
	}

	return nil
}

func (s *Scope) GetTypeOfChildNode(child *Node, objectType *ObjectType) *Type {
	switch child.Type {
	case NodeTypeIdentifier:
		return objectType.Properties[child.AsIdentifier().Name].Type
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

// MergeScope mean Adding other Locals to current Scope Locals
func (s *Scope) MergeScopeLocals(other *Scope) {
	for _, symbol := range other.Locals {
		s.AddVariable(symbol.Name, symbol.OriginalName, symbol.Type)
	}
}

type SymbolTable map[string]*Symbol

type TypeFlags uint32

const (
	TypeFlagsNone TypeFlags = 0

	TypeFlagsAny      TypeFlags = 1 << 0
	TypeFlagsObject   TypeFlags = 1 << 1
	TypeFlagsString   TypeFlags = 1 << 2
	TypeFlagsNumber   TypeFlags = 1 << 3
	TypeFlagsBoolean  TypeFlags = 1 << 4
	TypeFlagsNull     TypeFlags = 1 << 5
	TypeFlagsFunction TypeFlags = 1 << 6
)

func (t TypeFlags) String() string {
	switch t {
	case TypeFlagsNone:
		return "none"
	case TypeFlagsAny:
		return "any"
	case TypeFlagsObject:
		return "object"
	case TypeFlagsString:
		return "string"
	case TypeFlagsNumber:
		return "number"
	case TypeFlagsBoolean:
		return "boolean"
	case TypeFlagsNull:
		return "null"
	case TypeFlagsFunction:
		return "function"
	default:
		return "unknown"
	}
}

type Type struct {
	Flags TypeFlags
	Data  TypeData
}

func NewType[T TypeData](typeData T) T {
	_type := typeData.AsType()
	_type.Flags = typeData.Flags()
	_type.Data = typeData
	return typeData
}

type TypeData interface {
	AsType() *Type
	Flags() TypeFlags
	Name() string
}

func (t *Type) AsType() *Type                 { return t }
func (t *Type) AsStringType() *StringType     { return t.Data.(*StringType) }
func (t *Type) AsNumberType() *NumberType     { return t.Data.(*NumberType) }
func (t *Type) AsBooleanType() *BooleanType   { return t.Data.(*BooleanType) }
func (t *Type) AsNullType() *NullType         { return t.Data.(*NullType) }
func (t *Type) AsObjectType() *ObjectType     { return t.Data.(*ObjectType) }
func (t *Type) AsFunctionType() *FunctionType { return t.Data.(*FunctionType) }

type StringType struct {
	Type
}

func NewStringType() *StringType       { return &StringType{} }
func (t *StringType) Flags() TypeFlags { return TypeFlagsString }
func (t *StringType) Name() string     { return "string" }

type NumberType struct {
	Type
}

func NewNumberType() *NumberType       { return &NumberType{} }
func (t *NumberType) Flags() TypeFlags { return TypeFlagsNumber }
func (t *NumberType) Name() string     { return "number" }

type BooleanType struct {
	Type
}

func NewBooleanType() *BooleanType      { return &BooleanType{} }
func (t *BooleanType) Flags() TypeFlags { return TypeFlagsBoolean }
func (t *BooleanType) Name() string     { return "boolean" }

type NullType struct {
	Type
}

func NewNullType() *NullType         { return &NullType{} }
func (t *NullType) Flags() TypeFlags { return TypeFlagsNull }
func (t *NullType) Name() string     { return "null" }

type ObjectType struct {
	Type
	Properties map[string]*PropertyType
}

type PropertyType struct {
	Name         string
	OriginalName *string
	Type         *Type
}

func NewObjectType() *ObjectType       { return &ObjectType{Properties: make(map[string]*PropertyType)} }
func (t *ObjectType) Flags() TypeFlags { return TypeFlagsObject }
func (t *ObjectType) Name() string     { return "object" }
func (t *ObjectType) AddProperty(name string, propertyType *PropertyType) {
	t.Properties[name] = propertyType
}

type FunctionType struct {
	Type
	Params     []*Type
	ReturnType *Type
}

func NewFunctionType() *FunctionType             { return &FunctionType{Params: []*Type{}, ReturnType: &Type{}} }
func (t *FunctionType) Flags() TypeFlags         { return TypeFlagsFunction }
func (t *FunctionType) Name() string             { return "function" }
func (t *FunctionType) AddParamType(_type *Type) { t.Params = append(t.Params, _type) }

func InferTypeFromNode(node *Node) *Type {
	switch node.Type {
	case NodeTypeStringLiteral:
		return NewType(NewStringType()).AsType()
	case NodeTypeDecimalLiteral:
		return NewType(NewNumberType()).AsType()
	case NodeTypeBooleanLiteral:
		return NewType(NewBooleanType()).AsType()
	case NodeTypeNullLiteral:
		return NewType(NewNullType()).AsType()
	}

	return nil
}

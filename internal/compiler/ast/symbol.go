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

type SymbolTable map[string]*Symbol

type TypeFlags uint32

const (
	TypeFlagsNone TypeFlags = 0

	TypeFlagsAny     TypeFlags = 1 << 0
	TypeFlagsObject  TypeFlags = 1 << 1
	TypeFlagsString  TypeFlags = 1 << 2
	TypeFlagsNumber  TypeFlags = 1 << 3
	TypeFlagsBoolean TypeFlags = 1 << 4
	TypeFlagsNull    TypeFlags = 1 << 5
)

type Type struct {
	Flags TypeFlags
	Data  TypeData
}

type TypeData interface {
	Name() string
}

func (t *Type) AsStringType() *StringType   { return t.Data.(*StringType) }
func (t *Type) AsNumberType() *NumberType   { return t.Data.(*NumberType) }
func (t *Type) AsBooleanType() *BooleanType { return t.Data.(*BooleanType) }
func (t *Type) AsNullType() *NullType       { return t.Data.(*NullType) }
func (t *Type) AsObjectType() *ObjectType   { return t.Data.(*ObjectType) }

type StringType struct{}

func NewStringType() *StringType    { return &StringType{} }
func (s *StringType) ToType() *Type { return &Type{Data: s, Flags: TypeFlagsString} }
func (s *StringType) Name() string  { return "string" }

type NumberType struct{}

func NewNumberType() *NumberType    { return &NumberType{} }
func (s *NumberType) ToType() *Type { return &Type{Data: s, Flags: TypeFlagsNumber} }
func (s *NumberType) Name() string  { return "number" }

type BooleanType struct{}

func NewBooleanType() *BooleanType   { return &BooleanType{} }
func (s *BooleanType) ToType() *Type { return &Type{Data: s, Flags: TypeFlagsBoolean} }
func (s *BooleanType) Name() string  { return "boolean" }

type NullType struct{}

func NewNullType() *NullType      { return &NullType{} }
func (s *NullType) ToType() *Type { return &Type{Data: s, Flags: TypeFlagsNull} }
func (s *NullType) Name() string  { return "null" }

type ObjectType struct {
	Properties map[string]*PropertyType
}

type PropertyType struct {
	Name         string
	OriginalName *string
	Type         *Type
}

func NewObjectType() *ObjectType    { return &ObjectType{Properties: make(map[string]*PropertyType)} }
func (s *ObjectType) ToType() *Type { return &Type{Data: s, Flags: TypeFlagsObject} }
func (s *ObjectType) Name() string  { return "object" }

func (s *ObjectType) AddProperty(name string, propertyType *PropertyType) {
	s.Properties[name] = propertyType
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

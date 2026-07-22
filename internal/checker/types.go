package checker

import (
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/lexer"
	"fmt"
)

type TypeFlags uint32

const (
	TypeFlagsNone TypeFlags = 0

	TypeFlagsAny     TypeFlags = 1 << 0
	TypeFlagsObject  TypeFlags = 1 << 1
	TypeFlagsString  TypeFlags = 1 << 2
	TypeFlagsNumber  TypeFlags = 1 << 3
	TypeFlagsBoolean TypeFlags = 1 << 4
	TypeFlagsNull    TypeFlags = 1 << 5
	TypeFlagsUnion   TypeFlags = 1 << 7
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
	case TypeFlagsUnion:
		return "union"
	default:
		return "unknown"
	}
}

type ObjectFlags uint32

const (
	ObjectFlagsNone          ObjectFlags = 0
	ObjectFlagsInterface     ObjectFlags = 1 << 1 // Interface
	ObjectFlagsAnonymous     ObjectFlags = 1 << 2 // Anonymous
	ObjectFlagsObjectLiteral ObjectFlags = 1 << 3 // Originates in an object literal
	ObjectFlagsEvolvingArray ObjectFlags = 1 << 4 // Evolving array type
	ObjectFlagsArrayLiteral  ObjectFlags = 1 << 5 // Originates in an array literal
	ObjectFlagsReference     ObjectFlags = 1 << 6 // Generic type reference
)

func (o ObjectFlags) String() string {
	switch o {
	case ObjectFlagsNone:
		return "None"
	case ObjectFlagsInterface:
		return "Interface"
	case ObjectFlagsAnonymous:
		return "Anonymous"
	case ObjectFlagsObjectLiteral:
		return "ObjectLiteral"
	case ObjectFlagsEvolvingArray:
		return "EvolvingArray"
	case ObjectFlagsArrayLiteral:
		return "ArrayLiteral"
	default:
		return "unknown"
	}
}

type Type struct {
	Flags       TypeFlags
	ObjectFlags ObjectFlags
	Data        TypeData
	Symbol      *ast.Symbol
	Name        *string
}

type TypeData interface {
	AsType() *Type
	Name() string
}

func (t *Type) AsType() *Type                   { return t }
func (t *Type) AsIntrinsicType() *IntrinsicType { return t.Data.(*IntrinsicType) }
func (t *Type) AsObjectType() *ObjectType       { return t.Data.(*ObjectType) }
func (t *Type) AsUnionType() *UnionType         { return t.Data.(*UnionType) }

func (t *Type) ToString() string {
	switch v := t.Data.(type) {
	case *ObjectType:
		if v.signature != nil {
			return v.signature.ToString()
		}

		if v.ObjectFlags&ObjectFlagsInterface != 0 && t.Name != nil {
			return *t.Name
		}

		str := "{ "
		for name, member := range v.members {
			str += fmt.Sprintf("%s: %s؛ ", name, member.Type.ToString())
		}
		str += "}"
		return str
	case *IntrinsicType:
		switch {
		case v.Flags&TypeFlagsAny != 0:
			return lexer.TypeKeywordAny
		case v.Flags&TypeFlagsBoolean != 0:
			return lexer.TypeKeywordBoolean
		case v.Flags&TypeFlagsNumber != 0:
			return lexer.TypeKeywordNumber
		case v.Flags&TypeFlagsString != 0:
			return lexer.TypeKeywordString
		}
	case *UnionType:
		str := ""
		for index, _type := range v.types {
			str += _type.ToString()
			if index != len(v.types)-1 {
				str += " | "
			}
		}
		return str
	}

	return "؟"
}

type IntrinsicType struct {
	Type
	intrinsicName string
}

func (t *IntrinsicType) Name() string { return t.intrinsicName }

type LiteralType struct {
	Type
	value       string
	regularType *Type
}

func (t *LiteralType) Name() string { return t.value }

type ObjectTypeMember struct {
	Type         *Type
	OriginalName *string
}

type IndexInfo struct {
	keyType   *Type
	valueType *Type
}

type ObjectTypeMembers = map[string]*ObjectTypeMember
type ObjectType struct {
	Type
	members       ObjectTypeMembers
	signature     *Signature
	typeArguments map[string]*Type
	indexInfos    []*IndexInfo
}

func (t *ObjectType) Name() string               { return "object" }
func (t *ObjectType) Members() ObjectTypeMembers { return t.members }
func (t *ObjectType) Signature() *Signature      { return t.signature }

type UnionType struct {
	Type
	types []*Type
}

func NewUnionType(types []*Type) *UnionType { return &UnionType{types: types} }
func (t *UnionType) Name() string           { return "union" }

type SignatureFlags uint32

const (
	SignatureFlagsNone SignatureFlags = 0

	SignatureFlagsHasRestParameter SignatureFlags = 1 << 0 // Indicates last parameter is rest parameter
)

type SignatureParameter struct {
	Name string
	Type *Type
	Rest bool
}

type Signature struct {
	flags      SignatureFlags
	parameters []*SignatureParameter
	returnType *Type
	typeMapper map[string]*Type
}

func (s *Signature) Flags() SignatureFlags             { return s.flags }
func (s *Signature) Parameters() []*SignatureParameter { return s.parameters }
func (s *Signature) ReturnType() *Type                 { return s.returnType }

func (s *Signature) ToString() string {
	str := "("
	for index, param := range s.parameters {
		str += fmt.Sprintf("%s : %s", param.Name, param.Type.ToString())
		if index != len(s.parameters)-1 {
			str += ", "
		}
	}
	str += fmt.Sprintf(") => %s", s.returnType.ToString())
	return str
}

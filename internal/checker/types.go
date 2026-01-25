package checker

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
}

type TypeData interface {
	AsType() *Type
	Name() string
}

func (t *Type) AsType() *Type                   { return t }
func (t *Type) AsIntrinsicType() *IntrinsicType { return t.Data.(*IntrinsicType) }
func (t *Type) AsObjectType() *ObjectType       { return t.Data.(*ObjectType) }
func (t *Type) AsArrayType() *ArrayType         { return t.Data.(*ArrayType) }

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

type ArrayType struct {
	ObjectType
	ElementType *Type
}

func NewArrayType(elementType *Type) *ArrayType { return &ArrayType{ElementType: elementType} }
func (t *ArrayType) Name() string               { return "array" }

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
	Name   string
	Type   *Type
	isRest bool
}

type Signature struct {
	flags      SignatureFlags
	parameters []*SignatureParameter
	returnType *Type
}

package checker

import "arab_js/internal/compiler/ast"

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

func InferTypeFromNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeStringLiteral:
		return NewType(NewStringType()).AsType()
	case ast.NodeTypeDecimalLiteral:
		return NewType(NewNumberType()).AsType()
	case ast.NodeTypeBooleanLiteral:
		return NewType(NewBooleanType()).AsType()
	case ast.NodeTypeNullLiteral:
		return NewType(NewNullType()).AsType()
	}

	return nil
}

func GetTypeOfNode(scope *ast.Scope, node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeMemberExpression:
		memberExpression := node.AsMemberExpression()
		parent := GetTypeOfNode(scope, memberExpression.Object)
		return GetTypeOfChildNode(scope, memberExpression.Property, parent.AsObjectType())
	case ast.NodeTypeIdentifier:
		return GetTypeOfNode(scope, scope.GetVariableSymbol(node.AsIdentifier().Name).Node)
	}

	return nil
}

func GetTypeOfChildNode(scope *ast.Scope, child *ast.Node, objectType *ObjectType) *Type {
	switch child.Type {
	case ast.NodeTypeIdentifier:
		return objectType.Properties[child.AsIdentifier().Name].Type
	}

	return nil
}

func AreTypesCompatible(leftType *Type, rightType *Type) bool {
	if leftType == nil || rightType == nil {
		return false
	}

	if leftType.Flags == TypeFlagsAny || rightType.Flags == TypeFlagsAny {
		return true
	}

	if leftType.Flags == rightType.Flags {
		switch leftType.Flags {
		case TypeFlagsString, TypeFlagsNumber, TypeFlagsBoolean, TypeFlagsNull:
			return true
		}
	}

	if leftType.Flags == TypeFlagsObject && rightType.Flags == TypeFlagsObject {
		leftObj := leftType.AsObjectType()
		rightObj := rightType.AsObjectType()

		for name, rightProp := range rightObj.Properties {
			leftProp, ok := leftObj.Properties[name]
			if !ok {
				return false
			}
			if !AreTypesCompatible(leftProp.Type, rightProp.Type) {
				return false
			}
		}
		return true
	}

	if leftType.Flags == TypeFlagsFunction && rightType.Flags == TypeFlagsFunction {
		leftFn := leftType.AsFunctionType()
		rightFn := rightType.AsFunctionType()

		if len(leftFn.Params) != len(rightFn.Params) {
			return false
		}

		for i := range leftFn.Params {
			if !AreTypesCompatible(leftFn.Params[i], rightFn.Params[i]) {
				return false
			}
		}

		return AreTypesCompatible(leftFn.ReturnType, rightFn.ReturnType)
	}

	return false
}

package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
)

type TypeResolver struct {
	NameResolver *binder.NameResolver

	stringType  *Type
	numberType  *Type
	trueType    *Type
	falseType   *Type
	booleanType *Type
	nullType    *Type
	anyType     *Type
}

func NewTypeResolver(nameResolver *binder.NameResolver) *TypeResolver {
	t := &TypeResolver{
		NameResolver: nameResolver,
	}

	t.anyType = t.newIntrinsicType(TypeFlagsAny, "any")
	t.nullType = t.newIntrinsicType(TypeFlagsNull, "null")
	t.stringType = t.newIntrinsicType(TypeFlagsString, "string")
	t.numberType = t.newIntrinsicType(TypeFlagsNumber, "number")
	t.trueType = t.newLiteralType(TypeFlagsBoolean, "true")
	t.falseType = t.newLiteralType(TypeFlagsBoolean, "false")
	t.booleanType = t.newUnionType([]*Type{t.trueType, t.falseType})

	return t
}

func (t *TypeResolver) ResolveTypeFromNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeIdentifier:
		identifier := node.AsIdentifier()
		if node.Parent != nil && node.Parent.Type == ast.NodeTypeVariableDeclaration {
			return t.ResolveTypeAnnotation(identifier.TypeAnnotation)
		}

		if node.Parent != nil && node.Parent.Type == ast.NodeTypeFunctionDeclaration {
			return t.ResolveTypeAnnotation(identifier.TypeAnnotation)
		}

		symbol := t.NameResolver.Resolve(identifier.Name, node)
		if symbol == nil {
			return nil
		}

		return t.ResolveTypeFromNode(symbol.Node)
	case ast.NodeTypeVariableDeclaration:
		return t.ResolveTypeAnnotation(node.AsVariableDeclaration().Identifier.TypeAnnotation)
	case ast.NodeTypeFunctionDeclaration:
		return t.newFunctionType(t.resolveSignature(node))
	case ast.NodeTypeStringLiteral:
		return t.stringType
	case ast.NodeTypeDecimalLiteral:
		return t.numberType
	case ast.NodeTypeBooleanLiteral:
		return t.booleanType
	case ast.NodeTypeNullLiteral:
		return t.nullType
	case ast.NodeTypeMemberExpression:
		memberExpression := node.AsMemberExpression()
		objectType := t.ResolveTypeFromNode(memberExpression.Object)
		if objectType == nil {
			return nil
		}

		if objectType.Flags&TypeFlagsObject == 0 {
			return nil
		}

		return objectType.AsObjectType().members[memberExpression.PropertyName()].Type
	case ast.NodeTypeObjectExpression:
		objectExpression := node.AsObjectExpression()
		switch objectExpression.Parent.Type {
		case ast.NodeTypeInitializer:
			variableDeclaration := objectExpression.Parent.Parent.AsVariableDeclaration()
			return t.ResolveTypeAnnotation(variableDeclaration.Identifier.TypeAnnotation)
		case ast.NodeTypeAssignmentExpression:
			assignmentExpression := objectExpression.Parent.AsAssignmentExpression()
			return t.ResolveTypeFromNode(assignmentExpression.Left)
		default:
			members := map[string]*ObjectTypeMember{}
			for _, property := range objectExpression.Properties {
				if property.Type == ast.NodeTypeObjectProperty {
					objectProperty := property.AsObjectProperty()
					propertyName := objectProperty.Name()
					members[propertyName] = &ObjectTypeMember{
						Type:         t.inferTypeFromNode(objectProperty.Value),
						OriginalName: nil,
					}
				}
			}
			return t.newObjectLiteralType(members)
		}
	default:
		return nil
	}
}

func (t *TypeResolver) ResolveTypeAnnotation(typeAnnotation *ast.TypeAnnotation) *Type {
	if typeAnnotation == nil {
		return nil
	}

	return t.ResolveTypeNode(typeAnnotation.TypeAnnotation)
}

func (t *TypeResolver) ResolveTypeNode(typeNode *ast.Node) *Type {
	if typeNode == nil {
		return nil
	}

	switch typeNode.Type {
	case ast.NodeTypeStringKeyword:
		return t.stringType
	case ast.NodeTypeBooleanKeyword:
		return t.booleanType
	case ast.NodeTypeNullKeyword:
		return t.nullType
	case ast.NodeTypeNumberKeyword:
		return t.numberType
	case ast.NodeTypeAnyKeyword:
		return t.anyType
	case ast.NodeTypeTypeLiteral:
		typeLiteral := typeNode.AsTypeLiteral()
		members := map[string]*ObjectTypeMember{}
		for _, member := range typeLiteral.Members {
			propertySignature := member.AsPropertySignature()
			switch propertySignature.Key.Type {
			case ast.NodeTypeIdentifier:
				identifier := propertySignature.Key.AsIdentifier()
				members[identifier.Name] = &ObjectTypeMember{
					Type:         t.ResolveTypeNode(propertySignature.TypeNode()),
					OriginalName: identifier.OriginalName,
				}
			}
		}
		return t.newObjectLiteralType(members)
	case ast.NodeTypeTypeReference:
		return t.ResolveTypeFromTypeReference(typeNode.AsTypeReference())
	case ast.NodeTypeFunctionType:
		return t.newFunctionType(t.resolveSignature(typeNode))
	case ast.NodeTypeArrayType:
		arrayTypeNode := typeNode.AsArrayType()
		return t.newType(TypeFlagsObject, ObjectFlagsEvolvingArray, NewArrayType(t.ResolveTypeNode(arrayTypeNode.ElementType)))
	case ast.NodeTypeUnionType:
		unionTypeNode := typeNode.AsUnionType()
		types := []*Type{}
		for _, typeNode := range unionTypeNode.Types {
			types = append(types, t.ResolveTypeNode(typeNode))
		}
		return t.newUnionType(types)
	}

	return nil
}

func (t *TypeResolver) ResolveTypeFromTypeReference(typeReference *ast.TypeReference) *Type {
	symbol := t.NameResolver.Resolve(typeReference.TypeName.Name, typeReference.TypeName.AsNode())
	if symbol == nil {
		return nil
	}

	return t.ResolveTypeFromTypeDeclaration(symbol.Node)
}

func (t *TypeResolver) ResolveTypeFromTypeDeclaration(typeDeclaration *ast.Node) *Type {
	switch typeDeclaration.Type {
	case ast.NodeTypeInterfaceDeclaration:
		interfaceDeclaration := typeDeclaration.AsInterfaceDeclaration()
		interfaceType := t.newObjectType(ObjectFlagsInterface).AsObjectType()
		for _, member := range interfaceDeclaration.Body.Body {
			propertySignature := member.AsPropertySignature()
			switch propertySignature.Key.Type {
			case ast.NodeTypeIdentifier:
				identifier := propertySignature.Key.AsIdentifier()
				interfaceType.members[identifier.Name] = &ObjectTypeMember{
					Type:         t.ResolveTypeNode(propertySignature.TypeNode()),
					OriginalName: identifier.OriginalName,
				}
			}
		}
		return interfaceType.AsType()
	case ast.NodeTypeTypeAliasDeclaration:
		typeAliasDeclaration := typeDeclaration.AsTypeAliasDeclaration()
		return t.ResolveTypeAnnotation(typeAliasDeclaration.TypeAnnotation)
	default:
		return nil
	}
}

func (t *TypeResolver) Resolve(name string) *Type {
	symbol := t.NameResolver.Resolve(name, nil)

	if symbol != nil {
		return t.ResolveTypeFromTypeDeclaration(symbol.Node)
	}
	return nil
}

func (t *TypeResolver) GetApparentType(_type *Type) *Type {
	switch {
	case _type.Flags&TypeFlagsString != 0:
		return t.Resolve("تعريف_نص")
	default:
		return nil
	}
}

func (t *TypeResolver) newSignature(flags SignatureFlags, parameters []*SignatureParameter, returnType *Type) *Signature {
	return &Signature{
		flags:      flags,
		parameters: parameters,
		returnType: returnType,
	}
}

func (t *TypeResolver) newType(flags TypeFlags, objectFlags ObjectFlags, data TypeData) *Type {
	_type := data.AsType()
	_type.Flags = flags
	_type.ObjectFlags = objectFlags
	_type.Data = data
	return _type
}

func (t *TypeResolver) newIntrinsicType(flags TypeFlags, intrinsicName string) *Type {
	data := &IntrinsicType{}
	data.intrinsicName = intrinsicName
	return t.newType(flags, ObjectFlagsNone, data)
}

func (t *TypeResolver) newLiteralType(flags TypeFlags, value string) *Type {
	data := &LiteralType{}
	data.value = value
	data.regularType = data.AsType()
	return t.newType(flags, ObjectFlagsNone, data)
}

func (t *TypeResolver) newUnionType(types []*Type) *Type {
	data := &UnionType{}
	data.types = types
	return t.newType(TypeFlagsUnion, ObjectFlagsNone, data)
}

func (t *TypeResolver) newObjectType(objectFlags ObjectFlags) *Type {
	var data TypeData
	switch {
	case objectFlags&ObjectFlagsInterface != 0:
		data = &ObjectType{
			members: map[string]*ObjectTypeMember{},
		}
	case objectFlags&ObjectFlagsAnonymous != 0:
		data = &ObjectType{
			members: map[string]*ObjectTypeMember{},
		}
	default:
		panic("Unhandled case in newObjectType")
	}
	_type := t.newType(TypeFlagsObject, objectFlags, data)
	return _type
}

func (t *TypeResolver) newObjectLiteralType(members ObjectTypeMembers) *Type {
	return t.newType(TypeFlagsObject, ObjectFlagsObjectLiteral, &ObjectType{members: members})
}

func (t *TypeResolver) newFunctionType(signature *Signature) *Type {
	return t.newType(TypeFlagsObject, ObjectFlagsAnonymous, &ObjectType{signature: signature})
}

func (t *TypeResolver) resolveSignature(node *ast.Node) *Signature {
	switch node.Type {
	case ast.NodeTypeFunctionType:
		functionType := node.AsFunctionType()
		parameters := []*SignatureParameter{}
		flags := SignatureFlagsNone
		for _, param := range functionType.Params {
			isRest := false
			name := ""
			if param.Type == ast.NodeTypeRestElement {
				isRest = true
				name = param.AsRestElement().Argument.Name
				flags |= SignatureFlagsHasRestParameter
			} else {
				isRest = false
				name = param.AsIdentifier().Name
			}

			paramType := t.ResolveTypeFromNode(param)
			parameters = append(parameters, &SignatureParameter{Name: name, Type: paramType, isRest: isRest})
		}
		return t.newSignature(flags, parameters, t.ResolveTypeAnnotation(functionType.TypeAnnotation))
	case ast.NodeTypeFunctionDeclaration:
		functionDeclaration := node.AsFunctionDeclaration()
		parameters := []*SignatureParameter{}
		flags := SignatureFlagsNone
		for _, param := range functionDeclaration.Params {
			isRest := false
			name := ""
			if param.Type == ast.NodeTypeRestElement {
				isRest = true
				name = param.AsRestElement().Argument.Name
				flags |= SignatureFlagsHasRestParameter
			} else {
				isRest = false
				name = param.AsIdentifier().Name
			}

			paramType := t.ResolveTypeFromNode(param)
			parameters = append(parameters, &SignatureParameter{Name: name, Type: paramType, isRest: isRest})
		}
		return t.newSignature(flags, parameters, t.ResolveTypeAnnotation(functionDeclaration.TypeAnnotation))
	}
	return nil
}

func (t *TypeResolver) inferTypeFromNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeStringLiteral:
		return t.stringType
	case ast.NodeTypeDecimalLiteral:
		return t.numberType
	case ast.NodeTypeBooleanLiteral:
		return t.booleanType
	case ast.NodeTypeNullLiteral:
		return t.nullType
	case ast.NodeTypeObjectExpression:
		objectExpression := node.AsObjectExpression()
		members := map[string]*ObjectTypeMember{}
		for _, property := range objectExpression.Properties {
			switch property.Type {
			case ast.NodeTypeObjectProperty:
				objectProperty := property.AsObjectProperty()
				propertyName := objectProperty.Name()
				members[propertyName] = &ObjectTypeMember{
					Type:         t.inferTypeFromNode(objectProperty.Value),
					OriginalName: nil,
				}
			}
		}
		return t.newObjectLiteralType(members)
	}

	return nil
}

func (t *TypeResolver) areTypesCompatible(leftType *Type, rightType *Type) bool {
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

		for name, rightType := range rightObj.members {
			leftType, ok := leftObj.members[name]
			if !ok {
				return false
			}
			if !t.areTypesCompatible(leftType.Type, rightType.Type) {
				return false
			}
		}
		return true
	}

	if (leftType.Flags == TypeFlagsObject && leftType.AsObjectType().signature != nil) &&
		(rightType.Flags == TypeFlagsObject && rightType.AsObjectType().signature != nil) {
		leftFn := leftType.AsObjectType()
		rightFn := rightType.AsObjectType()

		// NOTE: Now we only support a single signature by function so this is correct.
		leftFnSignature := leftFn.signature
		rightFnSignature := rightFn.signature

		if len(leftFnSignature.parameters) != len(rightFnSignature.parameters) {
			return false
		}

		for i := range leftFnSignature.parameters {
			if !t.areTypesCompatible(
				leftFnSignature.parameters[i].Type,
				rightFnSignature.parameters[i].Type,
			) {
				return false
			}
		}

		return t.areTypesCompatible(leftFnSignature.returnType, rightFnSignature.returnType)
	}

	return false
}

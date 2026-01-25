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

	globalArrayType *Type
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

	t.globalArrayType = t.Resolve("المصفوفة")

	return t
}

func (t *TypeResolver) ResolveTypeFromNode(node *ast.Node) *Type {
	switch node.Type {
	case ast.NodeTypeParameter:
		return t.ResolveTypeAnnotation(node.AsParameter().TypeAnnotation)
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
		variableDeclaration := node.AsVariableDeclaration()
		if variableDeclaration.Name.Type == ast.NodeTypeIdentifier {
			return t.ResolveTypeAnnotation(variableDeclaration.Name.AsIdentifier().TypeAnnotation)
		}
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
			if variableDeclaration.Name.Type == ast.NodeTypeIdentifier {
				return t.ResolveTypeAnnotation(variableDeclaration.Name.AsIdentifier().TypeAnnotation)
			}
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

	return nil
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
		return t.ResolveTypeFromTypeReference(typeNode.AsTypeReferenceNode())
	case ast.NodeTypeFunctionType:
		return t.newFunctionType(t.resolveSignature(typeNode))
	case ast.NodeTypeArrayType:
		arrayTypeNode := typeNode.AsArrayType()
		return t.newArrayType(t.ResolveTypeNode(arrayTypeNode.ElementType))
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

func (t *TypeResolver) ResolveTypeFromTypeReference(typeReference *ast.TypeReferenceNode) *Type {
	symbol := t.NameResolver.Resolve(typeReference.TypeName.Name, typeReference.TypeName.AsNode())
	if symbol == nil {
		return nil
	}

	typeArguments := []*Type{}
	if typeReference.TypeParameters != nil {
		for _, typeParam := range typeReference.TypeParameters.Params {
			typeArguments = append(typeArguments, t.ResolveTypeNode(typeParam))

		}
	}

	return t.ResolveTypeFromTypeDeclaration(symbol.Node, typeArguments)
}

func (t *TypeResolver) ResolveProperty(typeNode *ast.Node, objectType *ObjectType) *Type {
	if typeNode.Type == ast.NodeTypeTypeReference {
		typeReference := typeNode.AsTypeReferenceNode()
		if _type, ok := objectType.typeArguments[typeReference.TypeName.Name]; ok {
			return _type
		}
		return t.ResolveTypeFromTypeReference(typeReference)
	}
	return t.ResolveTypeNode(typeNode)
}

func (t *TypeResolver) ResolveTypeFromTypeDeclaration(typeDeclaration *ast.Node, typesArguments []*Type) *Type {
	switch typeDeclaration.Type {
	case ast.NodeTypeInterfaceDeclaration:
		interfaceDeclaration := typeDeclaration.AsInterfaceDeclaration()
		interfaceType := t.newObjectType(ObjectFlagsInterface).AsObjectType()

		if interfaceDeclaration.TypeParameters != nil {
			for index, typeParameter := range interfaceDeclaration.TypeParameters.Params {
				interfaceType.typeArguments[typeParameter.Name] = typesArguments[index]
			}
		}

		for _, member := range interfaceDeclaration.Body.Body {
			switch member.Type {
			case ast.NodeTypePropertySignature:
				propertySignature := member.AsPropertySignature()
				switch propertySignature.Key.Type {
				case ast.NodeTypeIdentifier:
					identifier := propertySignature.Key.AsIdentifier()
					interfaceType.members[identifier.Name] = &ObjectTypeMember{
						Type:         t.ResolveProperty(propertySignature.TypeNode(), interfaceType),
						OriginalName: identifier.OriginalName,
					}
				}
			case ast.NodeTypeIndexSignatureDeclaration:
				indexSignatureDeclaration := member.AsIndexSignatureDeclaration()
				keyType := t.ResolveTypeAnnotation(indexSignatureDeclaration.Index.AsIdentifier().TypeAnnotation)
				valueType := t.ResolveTypeNode(indexSignatureDeclaration.Type)
				interfaceType.indexInfos = append(interfaceType.indexInfos, &IndexInfo{keyType: keyType, valueType: valueType})

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
		return t.ResolveTypeFromTypeDeclaration(symbol.Node, []*Type{})
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
			members:       map[string]*ObjectTypeMember{},
			indexInfos:    []*IndexInfo{},
			typeArguments: map[string]*Type{},
			signature:     nil,
		}
	case objectFlags&ObjectFlagsAnonymous != 0:
		data = &ObjectType{
			members:       map[string]*ObjectTypeMember{},
			indexInfos:    []*IndexInfo{},
			typeArguments: map[string]*Type{},
			signature:     nil,
		}
	default:
		panic("Unhandled case in newObjectType")
	}
	_type := t.newType(TypeFlagsObject, objectFlags, data)
	return _type
}

func (t *TypeResolver) newArrayType(elementType *Type) *Type {
	symbol := t.NameResolver.Resolve("المصفوفة", nil)
	if symbol == nil {
		return nil
	}

	return t.ResolveTypeFromTypeDeclaration(symbol.Node, []*Type{elementType})
}

func (t *TypeResolver) newObjectLiteralType(members ObjectTypeMembers) *Type {
	return t.newType(TypeFlagsObject, ObjectFlagsObjectLiteral, &ObjectType{
		members:       members,
		signature:     nil,
		typeArguments: map[string]*Type{},
		indexInfos:    []*IndexInfo{},
	})
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
			if param.Type == ast.NodeTypeParameter {
				parameter := param.AsParameter()
				if parameter.Name != nil {
					switch parameter.Name.Type {
					case ast.NodeTypeIdentifier:
						name = parameter.Name.AsIdentifier().Name
					}
				}

				isRest = parameter.Rest
				if isRest {
					flags |= SignatureFlagsHasRestParameter
				}

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
			if param.Type == ast.NodeTypeParameter {
				parameter := param.AsParameter()
				if parameter.Name != nil {
					switch parameter.Name.Type {
					case ast.NodeTypeIdentifier:
						name = parameter.Name.AsIdentifier().Name
					}
				}

				isRest = parameter.Rest
				if isRest {
					flags |= SignatureFlagsHasRestParameter
				}

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

func (t *TypeResolver) isTypeRelatedTo(target *Type, source *Type) bool {
	if t.isSimpleTypeRelatedTo(target, source) {
		return true
	}

	if target.Flags&TypeFlagsObject != 0 && source.Flags&TypeFlagsObject != 0 {
		targetType := target.AsObjectType()
		sourceType := source.AsObjectType()

		// FIXME: When adding optional members this will not be correct
		if len(targetType.members) != len(sourceType.members) {
			return false
		}

		for targetKey, targetMember := range targetType.members {
			sourceMember, ok := sourceType.members[targetKey]
			if !ok {
				return false
			}

			if !t.isTypeRelatedTo(targetMember.Type, sourceMember.Type) {
				return false
			}
		}

		if len(targetType.typeArguments) != len(sourceType.typeArguments) {
			return false
		}

		for index, targetTypeArgument := range targetType.typeArguments {
			sourceTypeArgument := sourceType.typeArguments[index]
			if !t.isTypeRelatedTo(targetTypeArgument, sourceTypeArgument) {
				return false
			}
		}

		return true
	}

	return false
}

func (t *TypeResolver) isSimpleTypeRelatedTo(target *Type, source *Type) bool {
	if target == nil || source == nil {
		return false
	}

	if target.Flags == TypeFlagsAny {
		return true
	}

	if target.Flags&TypeFlagsString != 0 && source.Flags&TypeFlagsString != 0 {
		return true
	}

	if target.Flags&TypeFlagsNumber != 0 && source.Flags&TypeFlagsNumber != 0 {
		return true
	}

	if target.Flags&TypeFlagsBoolean != 0 && source.Flags&TypeFlagsBoolean != 0 {
		return true
	}

	if target.Flags&TypeFlagsNull != 0 && source.Flags&TypeFlagsNull != 0 {
		return true
	}

	return false
}

func (t *TypeResolver) getPropertyOfType(_type *Type, name string) *Type {
	switch {
	case _type.Flags&TypeFlagsObject != 0:
		propertyType, ok := _type.AsObjectType().members[name]
		if ok {
			return propertyType.Type
		}
		return nil
	default:
		return nil
	}
}

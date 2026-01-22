package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
	"fmt"
)

type Program interface {
	SourceFiles() []*ast.SourceFile
	BindSourceFiles()
}

type Checker struct {
	program      Program
	NameResolver *binder.NameResolver
	TypeResolver *TypeResolver
	Diagnostics  []*ast.Diagnostic

	currentSourceFile *ast.SourceFile
}

func NewChecker(program Program) *Checker {
	program.BindSourceFiles()

	globalScope := &ast.Scope{}

	for _, sourceFile := range program.SourceFiles() {
		if !ast.IsExternalModule(sourceFile) {
			globalScope.MergeScopeLocals(sourceFile.Scope)
		}
	}

	nameResolver := binder.NewNameResolver(globalScope)

	return &Checker{
		program:      program,
		Diagnostics:  []*ast.Diagnostic{},
		NameResolver: nameResolver,
		TypeResolver: NewTypeResolver(nameResolver),
	}
}

func (c *Checker) error(location ast.Location, message string) {
	c.Diagnostics = append(c.Diagnostics,
		ast.NewDiagnostic(
			c.currentSourceFile,
			location,
			message,
		),
	)
}

func (c *Checker) errorf(location ast.Location, format string, a ...any) {
	c.error(location, fmt.Sprintf(format, a...))
}

func (c *Checker) Check() {
	for _, sourceFile := range c.program.SourceFiles() {
		c.currentSourceFile = sourceFile
		c.checkStatementList(sourceFile.Body)
	}
}

func (c *Checker) checkStatementList(nodes []*ast.Node) {
	for _, node := range nodes {
		c.checkStatement(node)
	}
}

func (c *Checker) checkBlockStatement(blockStatement *ast.BlockStatement) {
	c.checkStatementList(blockStatement.Body)
}

func (c *Checker) checkStatement(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeVariableDeclaration:
		c.checkVariableDeclaration(node.AsVariableDeclaration())
	case ast.NodeTypeExpressionStatement:
		c.checkExpression(node.AsExpressionStatement().Expression)
	case ast.NodeTypeBlockStatement:
		c.checkBlockStatement(node.AsBlockStatement())
	case ast.NodeTypeIfStatement:
		c.checkIfStatement(node.AsIfStatement())
	case ast.NodeTypeReturnStatement:
		c.checkReturnStatement(node.AsReturnStatement())
	}
}

func (c *Checker) checkReturnStatement(returnStatement *ast.ReturnStatement) {
	c.checkExpression(returnStatement.Argument)
}

func (c *Checker) checkIfStatement(ifStatement *ast.IfStatement) {
	c.checkExpression(ifStatement.TestExpression)
	c.checkStatement(ifStatement.ConsequentStatement)
	if ifStatement.AlternateStatement != nil {
		c.checkStatement(ifStatement.AlternateStatement)
	}
}

func (c *Checker) checkVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Name.Type == ast.NodeTypeIdentifier &&
		variableDeclaration.Name.AsIdentifier().TypeAnnotation == nil {
		// NOTE: Identifier is any so we skip
		return
	}

	identifierType := c.TypeResolver.ResolveTypeNode(variableDeclaration.TypeNode())

	if variableDeclaration.Initializer == nil {
		return
	}
	initializerType := c.checkExpression(variableDeclaration.Initializer.Expression)
	if initializerType == nil {
		return
	}

	if !c.TypeResolver.areTypesCompatible(identifierType, initializerType) {
		c.errorf(variableDeclaration.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, identifierType.Data.Name(), initializerType.Data.Name())
		return
	}
}

func (c *Checker) checkArrayExpression(arrayExpression *ast.ArrayExpression) *Type {
	elementTypes := []*Type{}
	for _, element := range arrayExpression.Elements {
		elementTypes = append(elementTypes, c.checkExpression(element))
	}
	elementType := c.TypeResolver.newUnionType(elementTypes)

	return c.TypeResolver.newType(TypeFlagsObject, ObjectFlagsArrayLiteral, NewArrayType(elementType))
}

func (c *Checker) checkExpression(expression *ast.Node) *Type {
	switch expression.Type {
	case ast.NodeTypeIdentifier:
		return c.checkIdentifier(expression.AsIdentifier())
	case
		ast.NodeTypeStringLiteral,
		ast.NodeTypeDecimalLiteral,
		ast.NodeTypeBooleanLiteral,
		ast.NodeTypeNullLiteral:
		return c.TypeResolver.ResolveTypeFromNode(expression)
	case ast.NodeTypeArrayExpression:
		return c.checkArrayExpression(expression.AsArrayExpression())
	case ast.NodeTypeAssignmentExpression:
		assignmentExpression := expression.AsAssignmentExpression()
		leftType := c.checkExpression(assignmentExpression.Left)
		rightType := c.checkExpression(assignmentExpression.Right)
		if leftType == nil {
			return nil
		}

		if rightType == nil {
			return nil
		}

		if !c.TypeResolver.areTypesCompatible(leftType, rightType) {
			c.errorf(assignmentExpression.Location, TYPE_0_IS_NOT_ASSIGNABLE_TO_TYPE_1_FORMAT, leftType.Data.Name(), rightType.Data.Name())
			return nil
		}
		return leftType
	case ast.NodeTypeCallExpression:
		callExpression := expression.AsCallExpression()
		return c.checkCallExpression(callExpression)
	case ast.NodeTypeObjectExpression:
		objectExpression := expression.AsObjectExpression()
		return c.checkObjectExpression(objectExpression)
	case ast.NodeTypeMemberExpression:
		memberExpression := expression.AsMemberExpression()
		return c.checkMemberExpression(memberExpression)
	}

	return nil
}

func (c *Checker) checkCallExpression(callExpression *ast.CallExpression) *Type {
	_type := c.checkExpression(callExpression.Callee)
	if _type == nil {
		return nil
	}

	if _type.Flags&TypeFlagsObject == 0 {
		c.errorf(callExpression.Location, THIS_EXPRESSION_IS_NOT_CALLABLE_TYPE_0_HAS_NO_CALL_SIGNATURES, _type.Data.Name())
		return nil
	}

	objectType := _type.AsObjectType()
	if objectType.signature == nil {
		c.errorf(callExpression.Location, THIS_EXPRESSION_IS_NOT_CALLABLE_TYPE_0_HAS_NO_CALL_SIGNATURES, _type.Data.Name())
		return nil
	}

	if objectType.signature.flags&SignatureFlagsHasRestParameter == 0 {
		if len(objectType.signature.parameters) != len(callExpression.Args) {
			c.errorf(callExpression.Location, EXPECTED_0_ARGUMENTS_BUT_GOT_1, len(objectType.signature.parameters), len(callExpression.Args))
			return nil
		}
	}

	var restType *Type = nil
	var restIndex = -1
	for index, param := range objectType.signature.parameters {
		if param.isRest {
			restIndex = index
			restType = param.Type
			break
		}

		arg := callExpression.Args[index]
		_type := c.checkExpression(arg)
		if !c.TypeResolver.areTypesCompatible(param.Type, _type) {
			c.errorf(callExpression.Location, ARGUMENT_OF_TYPE_0_IS_NOT_ASSIGNABLE_TO_PARAMETER_OF_TYPE_1, _type.Data.Name(), param.Name)
			return nil
		}
	}

	if restType != nil {
		for index := restIndex; index < len(callExpression.Args); index++ {
			param := objectType.signature.parameters[restIndex]
			restElementType := param.Type.AsArrayType().ElementType
			arg := callExpression.Args[index]
			_type := c.checkExpression(arg)
			if !c.TypeResolver.areTypesCompatible(restElementType, _type) {
				c.errorf(callExpression.Location, ARGUMENT_OF_TYPE_0_IS_NOT_ASSIGNABLE_TO_PARAMETER_OF_TYPE_1, _type.Data.Name(), param.Name)
				return nil
			}
		}
	}

	return objectType.signature.returnType
}

func (c *Checker) checkObjectExpression(objectExpression *ast.ObjectExpression) *Type {
	return c.TypeResolver.inferTypeFromNode(objectExpression.AsNode())
}

func (c *Checker) checkMemberExpression(memberExpression *ast.MemberExpression) *Type {
	objectType := c.checkExpression(memberExpression.Object)
	if objectType == nil {
		return nil
	}

	propertyName := memberExpression.PropertyName()

	if objectType.Flags&TypeFlagsObject != TypeFlagsObject {
		// FIXME: in JavaScript everything is object
		return nil
	}

	if objectType.ObjectFlags&ObjectFlagsEvolvingArray != 0 {
		return objectType.AsArrayType().ElementType
	}

	propertyType := objectType.AsObjectType().members[propertyName]
	if propertyType == nil {
		c.errorf(memberExpression.AsNode().Location, PROPERTY_0_DOES_NOT_EXIST_ON_TYPE_1, propertyName, objectType.Data.Name())
		return nil
	}
	return propertyType.Type
}

func (c *Checker) checkIdentifier(identifier *ast.Identifier) *Type {
	symbol := c.NameResolver.Resolve(identifier.Name, identifier.AsNode())
	if symbol == nil {
		c.errorf(identifier.Location, CANNOT_FIND_NAME_0, identifier.Name)
		return nil
	}

	return c.TypeResolver.ResolveTypeFromNode(symbol.Node)
}

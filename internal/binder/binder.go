package binder

import (
	"arab_js/internal/compiler/ast"
	"fmt"
)

type Binder struct {
	sourceFile *ast.SourceFile
	container  *ast.ContainerBase

	Diagnostics []*ast.Diagnostic
}

func NewBinder(sourceFile *ast.SourceFile) *Binder {
	return &Binder{
		sourceFile: sourceFile,
		container:  sourceFile.ContainerBaseData(),
	}
}

func BindSourceFile(sourceFile *ast.SourceFile) *Binder {
	b := NewBinder(sourceFile)
	b.Bind()
	return b
}

func (b *Binder) Bind() {
	b.sourceFile.Scope = &ast.Scope{
		Locals:   map[string]*ast.Symbol{},
		Parent:   nil,
		IsGlobal: true,
	}

	for _, node := range b.sourceFile.Body {
		b.bindStatement(node)
	}
}

func (b *Binder) bindStatement(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Type {
	case ast.NodeTypeIfStatement:
		ifStatement := node.AsIfStatement()
		b.bindStatement(ifStatement.ConsequentStatement)
		if ifStatement.AlternateStatement != nil {
			b.bindStatement(ifStatement.AlternateStatement)
		}
	case ast.NodeTypeBlockStatement:
		b.bindBlockStatement(node.AsBlockStatement())
	case ast.NodeTypeVariableStatement:
		b.bindVariableStatement(node.AsVariableStatement())
	case ast.NodeTypeInterfaceDeclaration:
		b.bindInterfaceDeclaration(node.AsInterfaceDeclaration())
	case ast.NodeTypeFunctionDeclaration:
		b.bindFunctionDeclaration(node.AsFunctionDeclaration())
	case ast.NodeTypeForStatement:
		b.bindForStatement(node.AsForStatement())
	case ast.NodeTypeExpressionStatement:
		b.bindExpression(node.AsExpressionStatement().Expression)
	}
}

func (b *Binder) bindVariableStatement(variableStatement *ast.VariableStatement) {
	if variableStatement.DeclarationList == nil {
		return
	}

	b.bindVariableDeclarationList(variableStatement.DeclarationList.AsVariableDeclarationList())
}

func (b *Binder) bindVariableDeclarationList(variableDeclarationList *ast.VariableDeclarationList) {
	for _, variableDeclaration := range variableDeclarationList.Declarations {
		b.bindVariableDeclaration(variableDeclaration.AsVariableDeclaration())
	}
}

func (b *Binder) bindVariableDeclaration(variableDeclaration *ast.VariableDeclaration) {
	if variableDeclaration.Name.Type == ast.NodeTypeIdentifier {
		variableDeclaration.Symbol = b.declareSymbol(
			b.container.Scope,
			variableDeclaration.Name.AsIdentifier().Name,
			variableDeclaration.Name.AsIdentifier().OriginalName,
			variableDeclaration.AsNode(),
			ast.SymbolFlagsBlockScopedVariable,
		)
	}
}

func (b *Binder) bindBlockStatement(blockStatement *ast.BlockStatement) {
	saveContainer := b.container

	if canCreateNewScope(blockStatement.AsNode()) {
		blockStatement.Scope = &ast.Scope{IsGlobal: false}
		blockStatement.Scope.Parent = b.container.Scope
		b.container = blockStatement.ContainerBaseData()
	}

	for _, node := range blockStatement.Body {
		b.bindStatement(node)
	}

	b.container = saveContainer
}

func (b *Binder) bindTypeParam(typeParameter *ast.TypeParameter) {
	b.declareSymbol(
		b.container.Scope,
		typeParameter.Name,
		nil,
		typeParameter.AsNode(),
		ast.SymbolFlagsTypeParameter,
	)
}

func (b *Binder) bindInterfaceDeclaration(interfaceDeclaration *ast.InterfaceDeclaration) {
	interfaceDeclaration.Symbol = b.declareSymbol(
		b.container.Scope,
		interfaceDeclaration.Id.Name,
		nil,
		interfaceDeclaration.AsNode(),
		ast.SymbolFlagsInterface,
	)

	saveContainer := b.container
	interfaceDeclaration.Scope = &ast.Scope{IsGlobal: false}
	interfaceDeclaration.Scope.Parent = b.container.Scope
	b.container = interfaceDeclaration.ContainerBaseData()

	if interfaceDeclaration.TypeParameters != nil {
		for _, param := range interfaceDeclaration.TypeParameters.Params {
			b.bindTypeParam(param)
		}
	}

	b.container = saveContainer
}

func (b *Binder) bindFunctionDeclaration(functionDeclaration *ast.FunctionDeclaration) {
	functionDeclaration.Symbol = b.declareSymbol(
		b.container.Scope,
		functionDeclaration.ID.Name,
		functionDeclaration.ID.OriginalName,
		functionDeclaration.AsNode(),
		ast.SymbolFlagsFunction,
	)

	saveContainer := b.container
	functionDeclaration.Scope = &ast.Scope{IsGlobal: false}
	functionDeclaration.Scope.Parent = b.container.Scope
	b.container = functionDeclaration.ContainerBaseData()

	for _, param := range functionDeclaration.Params {
		b.bindParam(param)
	}

	if functionDeclaration.Body != nil {
		b.bindBlockStatement(functionDeclaration.Body)
	}

	b.container = saveContainer
}

func (b *Binder) bindParam(node *ast.Node) {
	switch node.Type {
	case ast.NodeTypeParameter:
		param := node.AsParameter()
		if param.Name == nil {
			return
		}

		switch param.Name.Type {
		case ast.NodeTypeIdentifier:
			identifier := param.Name.AsIdentifier()
			b.declareSymbol(
				b.container.Scope,
				identifier.Name,
				nil,
				node,
				ast.SymbolFlagsFunctionScopedVariable,
			)
		}
	}
}

func (b *Binder) bindForStatement(forStatement *ast.ForStatement) {
	saveContainer := b.container
	forStatement.Scope = &ast.Scope{IsGlobal: false}
	forStatement.Scope.Parent = b.container.Scope
	b.container = forStatement.ContainerBaseData()

	switch forStatement.Init.Type {
	case ast.NodeTypeVariableStatement:
		b.bindVariableStatement(forStatement.Init.AsVariableStatement())
	}

	b.bindStatement(forStatement.Body)

	b.container = saveContainer
}

func (b *Binder) bindExpression(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Type {
	case ast.NodeTypeCallExpression:
		callExpression := node.AsCallExpression()
		b.bindExpression(callExpression.Callee)
		for _, arg := range callExpression.Args {
			b.bindExpression(arg)
		}
	case ast.NodeTypeArrowFunction:
		b.bindArrowFunction(node.AsArrowFunction())
	case ast.NodeTypeMemberExpression:
		memberExpression := node.AsMemberExpression()
		b.bindExpression(memberExpression.Object)
		b.bindExpression(memberExpression.Property)
	case ast.NodeTypeAssignmentExpression:
		b.bindExpression(node.AsAssignmentExpression().Left)
		b.bindExpression(node.AsAssignmentExpression().Right)
	}
}

func (b *Binder) bindArrowFunction(arrowFunction *ast.ArrowFunction) {
	saveContainer := b.container
	arrowFunction.Scope = &ast.Scope{IsGlobal: false}
	arrowFunction.Scope.Parent = b.container.Scope
	b.container = arrowFunction.ContainerBaseData()

	for _, param := range arrowFunction.Params {
		b.bindParam(param)
	}

	if arrowFunction.Body == nil {
		return
	}

	if arrowFunction.Body.Type == ast.NodeTypeBlockStatement {
		b.bindBlockStatement(arrowFunction.Body.AsBlockStatement())
	}

	b.container = saveContainer
}

func (b *Binder) declareSymbol(
	scope *ast.Scope,
	name string,
	originalName *string,
	node *ast.Node,
	flags ast.SymbolFlags,
) *ast.Symbol {
	if scope.Locals == nil {
		scope.Locals = make(map[string]*ast.Symbol)
	}

	if existingSymbol, exists := scope.Locals[name]; exists {
		if existingSymbol.Flags&ast.SymbolFlagsBlockScoped != 0 {
			b.errorf(node.DeclarationBaseData().IdentifierNameNode().Location, CANNOT_REDECLARE_BLOCK_SCOPED_VARIABLE_0, name)
		} else {
			b.errorf(node.DeclarationBaseData().IdentifierNameNode().Location, DUPLICATE_IDENTIFIER_0, name)
		}
		return nil
	}

	return scope.AddVariable(name, originalName, node, flags)
}

func (b *Binder) error(location ast.Location, message string) {
	b.Diagnostics = append(b.Diagnostics,
		ast.NewDiagnostic(
			b.sourceFile,
			location,
			message,
		),
	)
}

func (b *Binder) errorf(location ast.Location, format string, a ...any) {
	b.error(location, fmt.Sprintf(format, a...))
}

func canCreateNewScope(node *ast.Node) bool {
	if node.Parent == nil {
		return true
	}

	switch node.Parent.Type {
	case ast.NodeTypeFunctionDeclaration, ast.NodeTypeForStatement:
		return false
	default:
		return true
	}
}

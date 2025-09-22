package binder

import (
	"arab_js/internal/compiler/ast"
)

type NameResolver struct {
	Globals *ast.Scope
}

func NewNameResolver(globals *ast.Scope) *NameResolver {
	return &NameResolver{
		Globals: globals,
	}
}

func (n *NameResolver) Resolve(name string, location *ast.Node) *ast.Symbol {
	if location == nil && n.Globals != nil {
		return n.Globals.GetVariableSymbol(name)
	}

	currentScope := getCurrentScope(location)

	for currentScope != nil {
		if symbol := currentScope.GetVariableSymbol(name); symbol != nil {
			return symbol
		}
		currentScope = currentScope.Parent
	}

	if n.Globals != nil {
		return n.Globals.GetVariableSymbol(name)
	}

	return nil
}

func getCurrentScope(node *ast.Node) *ast.Scope {
	scope := node.LocalScope()
	for scope == nil {
		node = node.Parent
		if node == nil {
			return nil
		}
		scope = node.LocalScope()
	}
	return scope
}

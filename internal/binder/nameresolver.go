package binder

import "arab_js/internal/compiler/ast"

type NameResolver struct {
	Globals *ast.Scope
}

func NewNameResolver(globals *ast.Scope) *NameResolver {
	return &NameResolver{
		Globals: globals,
	}
}

func (n *NameResolver) Resolve(name string, scope *ast.Scope) *ast.Symbol {
	currentScope := scope

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

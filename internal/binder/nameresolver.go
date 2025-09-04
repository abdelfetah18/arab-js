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
	symbol := currentScope.GetVariableSymbol(name)
	if symbol == nil {
		currentScope = scope.Parent
		for currentScope != nil {
			symbol = currentScope.GetVariableSymbol(name)
			if symbol == nil {
				currentScope = scope.Parent
			} else {
				return symbol
			}
		}

		symbol = n.Globals.GetVariableSymbol(name)
	}

	return symbol
}

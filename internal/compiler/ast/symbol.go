package ast

type Symbol struct {
	Name         string
	OriginalName *string
	Node         *Node
	Flags        SymbolFlags
}

type Scope struct {
	Locals   map[string]*Symbol
	Parent   *Scope
	IsGlobal bool
}

func (s *Scope) AddVariable(name string, originalName *string, node *Node, flags SymbolFlags) *Symbol {
	symbol := Symbol{
		Name:         name,
		Node:         node,
		OriginalName: originalName,
		Flags:        flags,
	}
	s.Locals[name] = &symbol
	return &symbol
}

func (s *Scope) GetVariableSymbol(name string) *Symbol {
	if s.Locals == nil {
		return nil
	}

	if symbol, exists := s.Locals[name]; exists {
		return symbol
	}

	return nil
}

// MergeScope mean Adding other Locals to current Scope Locals
func (s *Scope) MergeScopeLocals(other *Scope, reportMergeSymbolError func(target *Symbol, source *Symbol)) {
	for _, symbol := range other.Locals {
		if s.Locals == nil {
			s.Locals = make(map[string]*Symbol)
		}

		if existingSymbol, exists := s.Locals[symbol.Name]; exists {
			reportMergeSymbolError(existingSymbol, symbol)
			continue
		}

		s.AddVariable(symbol.Name, symbol.OriginalName, symbol.Node, symbol.Flags)
	}
}

type SymbolTable map[string]*Symbol

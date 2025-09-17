package ast

type Symbol struct {
	Name         string
	OriginalName *string
	Node         *Node
}

type Scope struct {
	Locals map[string]*Symbol
	Parent *Scope
}

func (s *Scope) AddVariable(name string, originalName *string, node *Node) *Symbol {
	if s.Locals == nil {
		s.Locals = make(map[string]*Symbol)
	}

	if _, exists := s.Locals[name]; exists {
		return nil
	}

	symbol := Symbol{
		Name:         name,
		Node:         node,
		OriginalName: originalName,
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
func (s *Scope) MergeScopeLocals(other *Scope) {
	for _, symbol := range other.Locals {
		s.AddVariable(symbol.Name, symbol.OriginalName, symbol.Node)
	}
}

type SymbolTable map[string]*Symbol

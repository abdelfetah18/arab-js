package transformer

import (
	"arab_js/internal/compiler/ast"
)

type Transformer struct {
	program     *ast.Program
	symbolTable *ast.SymbolTable
}

func NewTransformer(program *ast.Program, symbolTable *ast.SymbolTable) *Transformer {
	return &Transformer{
		program:     program,
		symbolTable: symbolTable,
	}
}

func (t *Transformer) Transform() *ast.Node {
	currentScope := t.symbolTable.Current
	nodeVisitor := ast.NewNodeVisitor(func(node *ast.Node) *ast.Node {
		if node.Type == ast.NodeTypeBlockStatement {
			currentScope = node.AsBlockStatement().Scope
		}

		if node.Type == ast.NodeTypeIdentifier {
			identifier := node.AsIdentifier()
			symbol := currentScope.GetVariableSymbol(identifier.Name)
			if symbol != nil && symbol.OriginalName != nil {
				identifier.Name = *symbol.OriginalName
			}
		}

		return node
	})

	return nodeVisitor.VisitNode(t.program.ToNode())
}

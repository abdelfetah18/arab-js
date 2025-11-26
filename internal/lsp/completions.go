package lsp

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func getNodeAtPosition(sourceFile *ast.SourceFile, position uint) *ast.Node {
	var findNode *ast.Node = nil
	var current *ast.Node = sourceFile.AsNode()
	var next *ast.Node = nil

	visitAll := func(n *ast.Node) bool {
		if position >= n.Location.Pos && position <= n.Location.End {
			next = n
			return true
		}

		return false
	}

	for {
		found := current.ForEachChild(visitAll)
		if next == nil && !found {
			findNode = current
			break
		}

		current = next
		next = nil
	}

	if findNode != nil && findNode.Type == ast.NodeTypeMemberExpression {
		memberExpression := findNode.AsMemberExpression()
		propertyNode := memberExpression.Property
		if position >= propertyNode.Location.Pos && position <= propertyNode.Location.End {
			return propertyNode
		}
	}

	return findNode
}

func getCompletionData(node *ast.Node, checker *checker.Checker) []defines.CompletionItem {
	completions := []defines.CompletionItem{}

	isPropertyAccess := node.Parent != nil && node.Parent.Type == ast.NodeTypeMemberExpression
	if isPropertyAccess {
		_type := checker.TypeResolver.ResolveTypeFromNode(node.Parent.AsMemberExpression().Object)
		objectType := _type.AsObjectType()
		for name := range objectType.Members() {
			d := defines.CompletionItemKindText
			completions = append(completions, defines.CompletionItem{
				Label:      name,
				Kind:       &d,
				InsertText: &name,
			})
		}
	} else {
		var currentScope *ast.Scope = node.GetPrentContainer()

		for currentScope != nil {
			for k, symbol := range currentScope.Locals {
				d := defines.CompletionItemKindText
				switch symbol.Node.Type {
				case ast.NodeTypeFunctionDeclaration:
					d = defines.CompletionItemKindFunction
				case ast.NodeTypeVariableDeclaration:
					d = defines.CompletionItemKindVariable
				}
				completions = append(completions, defines.CompletionItem{
					Label:      k,
					Kind:       &d,
					InsertText: &k,
				})
			}
			currentScope = currentScope.Parent
		}

		for k, symbol := range checker.NameResolver.Globals.Locals {
			d := defines.CompletionItemKindText
			switch symbol.Node.Type {
			case ast.NodeTypeFunctionDeclaration:
				d = defines.CompletionItemKindFunction
			case ast.NodeTypeVariableDeclaration:
				d = defines.CompletionItemKindVariable
			}
			completions = append(completions, defines.CompletionItem{
				Label:      k,
				Kind:       &d,
				InsertText: &k,
			})
		}

	}

	return completions
}

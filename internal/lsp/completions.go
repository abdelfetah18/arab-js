package lsp

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/lexer"
	"math"
	"slices"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func getNodeAtPosition(sourceFile *ast.SourceFile, position uint) *ast.Node {
	var findNode *ast.Node = nil
	var current *ast.Node = sourceFile.AsNode()
	var next *ast.Node = nil

	smallest := math.MaxInt
	visitAll := func(n *ast.Node) bool {
		distance := n.Location.End - n.Location.Pos
		if position >= n.Location.Pos &&
			position <= n.Location.End &&
			distance <= uint(smallest) {
			next = n
			smallest = int(distance)
			return false
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

	return findNode
}

func getPrecedingTokensAtPos(sourceFile *ast.SourceFile, startPos int, position int) (contextToken lexer.Token, previousToken lexer.Token) {
	pos := startPos
	tokens := []lexer.Token{}
	lex := lexer.NewLexerAtPosition(sourceFile.Text, int(startPos))
	for pos < position {
		token := lex.Peek()
		tokens = append(tokens, token)
		pos = lex.Position()
		lex.Next()
	}

	previousToken = tokens[len(tokens)-1]
	if len(tokens) > 1 && previousToken.Type == lexer.Identifier {
		contextToken = tokens[len(tokens)-2]
		return contextToken, previousToken
	}

	return previousToken, previousToken
}

func getCompletionsFromSourceFile(sourceFile *ast.SourceFile, _checker *checker.Checker) []defines.CompletionItem {
	completions := []defines.CompletionItem{}
	var currentScope *ast.Scope = sourceFile.GetPrentContainer()

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

	for k, symbol := range _checker.NameResolver.Globals.Locals {
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

	return completions
}

func getCompletionsFromType(_type *checker.Type, propertiesToExclude []string) []defines.CompletionItem {
	completions := []defines.CompletionItem{}
	objectType := _type.AsObjectType()
	for name, members := range objectType.Members() {
		if slices.Contains(propertiesToExclude, name) {
			continue
		}
		d := defines.CompletionItemKindField
		if members.Type.ObjectFlags&checker.ObjectFlagsAnonymous != 0 && members.Type.AsObjectType().Signature() != nil {
			d = defines.CompletionItemKindMethod
		}
		completions = append(completions, defines.CompletionItem{
			Label:      name,
			Kind:       &d,
			InsertText: &name,
		})
	}
	return completions
}

func getCompletionData(sourceFile *ast.SourceFile, node *ast.Node, position int, _checker *checker.Checker) []defines.CompletionItem {
	contextNode := node.Parent
	completions := []defines.CompletionItem{}

	if (node.Type == ast.NodeTypeSourceFile) || (node.Type == ast.NodeTypeIdentifier && contextNode.Type == ast.NodeTypeExpressionStatement) {
		return getCompletionsFromSourceFile(sourceFile, _checker)
	}

	getCompletions := func(node *ast.Node, propertiesToExclude []string) []defines.CompletionItem {
		return getCompletionsFromType(_checker.TypeResolver.ResolveTypeFromNode(node), propertiesToExclude)
	}

	if contextNode.Type == ast.NodeTypeMemberExpression {
		return getCompletions(contextNode.AsMemberExpression().Object, []string{})
	}

	if node.Type == ast.NodeTypeObjectExpression {
		properties := node.AsObjectExpression().PropertiesNames()
		if contextNode.Type == ast.NodeTypeAssignmentExpression {
			return getCompletions(contextNode.AsAssignmentExpression().Left, properties)
		}
		if contextNode.Type == ast.NodeTypeInitializer {
			return getCompletions(contextNode.AsInitializer().Parent, properties)
		}
	}

	return completions
}

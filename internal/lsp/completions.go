package lsp

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/lexer"
	"math"
	"slices"
	"sync"

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

func getLeafNodeAtPosition(sourceFile *ast.SourceFile, position uint) *ast.Node {
	var curNode *ast.Node = nil

	visitNode := func(n *ast.Node) bool {
		if position >= n.Location.Pos &&
			position <= n.Location.End {
			curNode = n
			return true
		}

		return false
	}

	node := sourceFile.AsNode()
	for {
		node.ForEachChild(visitNode)
		if curNode == node {
			break
		}

		if curNode == nil {
			break
		}

		node = curNode
	}

	return curNode
}

func getPrecedingTokensAtPos(sourceFile *ast.SourceFile, startPos int, position int) (contextToken lexer.Token, previousToken lexer.Token) {
	tokens := []lexer.Token{}
	lex := lexer.NewLexerAtPosition(sourceFile.Text, int(startPos))

	token := lex.Peek()
	sPos := token.Position
	ePos := token.Position + len(token.Value)

	nextToken := func() {
		lex.Next()
		token = lex.Peek()
		sPos = token.Position
		ePos = sPos + len(lex.Peek().Value)
	}

	for {
		if ePos <= position {
			tokens = append(tokens, token)
		} else {
			break
		}
		nextToken()
		if token.Type == lexer.EOF || token.Type == lexer.Invalid {
			break
		}
	}

	previousToken = tokens[len(tokens)-1]
	if len(tokens) > 1 && previousToken.Type == lexer.Identifier {
		contextToken = tokens[len(tokens)-2]
		return contextToken, previousToken
	}

	return previousToken, previousToken
}

func getCompletionsFromScope(scope *ast.Scope, onlyTypes bool) []defines.CompletionItem {
	completions := []defines.CompletionItem{}
	for k, symbol := range scope.Locals {
		if onlyTypes && !slices.Contains(
			[]ast.NodeType{
				ast.NodeTypeInterfaceDeclaration,
				ast.NodeTypeTypeLiteral,
			},
			symbol.Node.Type,
		) {
			continue
		}
		d := defines.CompletionItemKindText
		switch symbol.Node.Type {
		case ast.NodeTypeFunctionDeclaration:
			d = defines.CompletionItemKindFunction
		case ast.NodeTypeVariableDeclaration:
			d = defines.CompletionItemKindVariable
		case ast.NodeTypeInterfaceDeclaration:
			d = defines.CompletionItemKindInterface
		}
		completions = append(completions, defines.CompletionItem{
			Label:      k,
			Kind:       &d,
			InsertText: &k,
		})
	}
	return completions
}

func getCompletionsFromCurrentNode(node *ast.Node, onlyTypes bool, _checker *checker.Checker) []defines.CompletionItem {
	completions := []defines.CompletionItem{}
	var currentScope *ast.Scope = node.GetParentContainer()

	for currentScope != nil {
		if currentScope.IsGlobal {
			break
		}
		completions = append(completions, getCompletionsFromScope(currentScope, onlyTypes)...)
		currentScope = currentScope.Parent
	}

	if !onlyTypes {
		completions = append(completions, allKeywordCompletions()...)
	}

	completions = append(completions, getCompletionsFromScope(_checker.NameResolver.Globals, onlyTypes)...)

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

func getCompletionData(sourceFile *ast.SourceFile, position int, _checker *checker.Checker) []defines.CompletionItem {
	completions := []defines.CompletionItem{}
	getCompletions := func(node *ast.Node, propertiesToExclude []string) []defines.CompletionItem {
		return getCompletionsFromType(_checker.TypeResolver.ResolveTypeFromNode(node, map[string]*checker.Type{}), propertiesToExclude)
	}

	previousNode := getPreviousNodeAtPosition(sourceFile, uint(position))
	contextToken, _ := getPrecedingTokensAtPos(sourceFile, int(previousNode.Location.Pos), position)

	if contextToken.Type == lexer.Dot {
		if previousNode.Parent.Type == ast.NodeTypeMemberExpression {
			return getCompletions(previousNode.Parent.AsMemberExpression().Object, []string{})
		}
	} else {
		// TypeOnly
		isTypeOnly := contextToken.Type == lexer.Colon && previousNode.Parent.Type != ast.NodeTypeObjectProperty
		if isTypeOnly {
			// prmitives + interfaces + types
			return append(types(), getCompletionsFromCurrentNode(previousNode, true, _checker)...)
		}

		node := getLeafNodeAtPosition(sourceFile, uint(position))
		// Object Property Name
		if (contextToken.Type == lexer.LeftCurlyBrace || contextToken.Type == lexer.Comma) && node.Type == ast.NodeTypeObjectExpression {
			contextNode := node.Parent
			properties := node.AsObjectExpression().PropertiesNames()
			if contextNode.Type == ast.NodeTypeAssignmentExpression {
				return getCompletions(contextNode.AsAssignmentExpression().Left, properties)
			}
			if contextNode.Type == ast.NodeTypeInitializer {
				return getCompletions(contextNode.AsInitializer().Parent, properties)
			}
		}

		// Value
		// from current scope to global scope
		return getCompletionsFromCurrentNode(node, false, _checker)
	}

	return completions
}

var (
	allKeywordCompletions = sync.OnceValue(func() []defines.CompletionItem {
		result := make([]defines.CompletionItem, 0, len(lexer.Keywords))

		for _, keword := range lexer.Keywords {
			d := defines.CompletionItemKindKeyword
			k := keword
			result = append(result, defines.CompletionItem{
				Label:      k,
				Kind:       &d,
				InsertText: &k,
			})
		}

		for _, typeKeword := range lexer.TypeKeywords {
			d := defines.CompletionItemKindKeyword
			k := typeKeword
			result = append(result, defines.CompletionItem{
				Label:      k,
				Kind:       &d,
				InsertText: &k,
			})
		}

		return result
	})
	types = sync.OnceValue(func() []defines.CompletionItem {
		types := []lexer.TypeKeyword{
			lexer.TypeKeywordAny,
			lexer.TypeKeywordString,
			lexer.TypeKeywordNumber,
			lexer.TypeKeywordBoolean,
		}
		result := make([]defines.CompletionItem, 0, len(types))

		for _, _type := range types {
			d := defines.CompletionItemKindKeyword
			k := _type
			result = append(result, defines.CompletionItem{
				Label:      k,
				Kind:       &d,
				InsertText: &k,
			})
		}

		return result
	})
)

func getMostRightNode(node *ast.Node, position uint) *ast.Node {
	smallest := math.MaxInt
	foundNode := node
	visitNode := func(n *ast.Node) bool {
		if (n.Location.End - n.Location.Pos) == 0 {
			return false
		}

		distance := position - n.Location.End
		if distance < uint(smallest) {
			foundNode = n
		}
		return false
	}

	curNode := node
	for {
		curNode.ForEachChild(visitNode)
		if foundNode == curNode {
			break
		}
		curNode = foundNode
	}
	return foundNode
}

func getPreviousNodeAtPosition(sourceFile *ast.SourceFile, position uint) *ast.Node {
	var curNode *ast.Node = nil
	var prevNode *ast.Node = nil

	visitNode := func(n *ast.Node) bool {
		if (n.Location.End - n.Location.Pos) == 0 {
			return false
		}

		if position >= n.Location.Pos &&
			position <= n.Location.End {
			curNode = n
			return true
		}

		if position >= n.Location.End {
			prevNode = n
		}

		return false
	}

	node := sourceFile.AsNode()
	for {
		node.ForEachChild(visitNode)
		if curNode == node {
			break
		}

		if curNode == nil {
			break
		}

		node = curNode
	}

	return getMostRightNode(prevNode, uint(position))
}

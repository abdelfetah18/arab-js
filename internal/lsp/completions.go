package lsp

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/lexer"
	"math"

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
			distance != 0 &&
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

func getPrecedingTokensAtPos(sourceFile *ast.SourceFile, startPos int, position int) (lexer.Token, lexer.Token) {
	pos := startPos
	tokens := []lexer.Token{}
	lex := lexer.NewLexerAtPosition(sourceFile.Text, int(startPos))
	for pos < position {
		token := lex.Peek()
		tokens = append(tokens, token)
		pos = lex.Position()
		lex.Next()
	}

	currentToken := tokens[len(tokens)-1]
	if len(tokens) > 1 {
		return tokens[len(tokens)-2], currentToken
	}

	return currentToken, currentToken
}

func getCompletionData(sourceFile *ast.SourceFile, node *ast.Node, position int, _checker *checker.Checker) []defines.CompletionItem {
	completions := []defines.CompletionItem{}

	if node.Type == ast.NodeTypeSourceFile {
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

	_, currentToken := getPrecedingTokensAtPos(sourceFile, int(node.Location.Pos), position)

	isRightOfDot := currentToken.Type == lexer.Dot
	if isRightOfDot {
		if node.Type == ast.NodeTypeMemberExpression {
			_type := _checker.TypeResolver.ResolveTypeFromNode(node.AsMemberExpression().Object)
			objectType := _type.AsObjectType()
			for name, members := range objectType.Members() {
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
		}
	}

	return completions
}

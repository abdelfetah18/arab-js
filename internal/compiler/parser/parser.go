package parser

import (
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/lexer"
	"arab_js/internal/stack"
	"fmt"
	"strings"
)

type ParserState struct {
	lexerState        lexer.LexerState
	startPositionsLen int
	diagnosticLen     int
}

type Parser struct {
	lexer          *lexer.Lexer
	startPositions stack.Stack[uint]

	Diagnostics []*ast.Diagnostic

	opts ast.SourceFileParseOptions

	hasPrecedingOriginalNameDirective bool
	originalNameDirectiveValue        string
}

func NewParser(lexer *lexer.Lexer, opts ast.SourceFileParseOptions) *Parser {
	return &Parser{
		lexer:          lexer,
		startPositions: stack.Stack[uint]{},
		Diagnostics:    []*ast.Diagnostic{},
		opts:           opts,
	}
}

func ParseSourceFile(sourceText string) *ast.SourceFile {
	return NewParser(lexer.NewLexer(sourceText), ast.SourceFileParseOptions{FileName: "", Path: ""}).Parse()
}

func (p *Parser) Parse() *ast.SourceFile {
	isDeclarationFile := strings.Contains(p.opts.FileName, ".d.") || strings.Contains(p.opts.FileName, ".تعريف.")
	p.markStartPosition()

	directives := []*ast.Directive{} // p.parseDirectives()

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != lexer.EOF && p.lexer.Peek().Type != lexer.Invalid {
		statement := p.parseStatement()
		p.optional(lexer.Semicolon)
		statements = append(statements, statement)
	}

	return ast.NewNode(ast.NewSourceFile(p.lexer.Text(), statements, directives, isDeclarationFile), ast.Location{Pos: p.startPositions.Pop(), End: uint(p.lexer.Position())})
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordIf)
	p.expected(lexer.LeftParenthesis)
	testExpression := p.parseExpression()
	p.expected(lexer.RightParenthesis)
	consequentStatement := p.parseStatement()

	if p.optionalKeyword(lexer.KeywordElse) {
		alternateStatement := p.parseStatement()
		return ast.NewNode(
			ast.NewIfStatement(testExpression, consequentStatement, alternateStatement),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		)
	}

	return ast.NewNode(
		ast.NewIfStatement(testExpression, consequentStatement, nil),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseLabelledStatement() *ast.LabelledStatement {
	p.markStartPosition()

	label := p.parseIdentifier(false).AsNode()
	p.expected(lexer.Colon)
	var statement *ast.Node = nil
	if p.lexer.Peek().Type == lexer.KeywordToken && p.lexer.Peek().Value == lexer.KeywordFunction {
		// FIXME: Implement parseModifiers method
		statement = p.parseFunctionDeclaration(&ast.ModifierList{}).AsNode()
	} else {
		statement = p.parseStatement()
	}

	return ast.NewNode(
		ast.NewLabelledStatement(label, statement),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseStatement() *ast.Node {
	token := p.lexer.Peek()

	switch token.Type {
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordIf:
			return p.parseIfStatement().AsNode()
		case lexer.KeywordLet, lexer.KeywordConst:
			state := p.mark()
			p.lexer.Next()
			if p.isIdentifier() || p.lexer.Peek().Type == lexer.LeftCurlyBrace || p.lexer.Peek().Type == lexer.LeftSquareBracket {
				p.rewind(state)
				return p.parseVariableStatement(nil).AsNode()
			}
			p.rewind(state)
		case lexer.KeywordFunction:
			return p.parseFunctionDeclaration(nil).AsNode()
		case lexer.KeywordImport:
			return p.parseImportDeclaration().AsNode()
		case lexer.KeywordExport:
			return p.parseExportDeclarationOrExportAssignment()
		case lexer.KeywordReturn:
			return p.parseReturnStatement().AsNode()
		case lexer.KeywordFor:
			return p.parseBreakableStatement()
		}
	case lexer.LeftCurlyBrace:
		return p.parseBlockStatement().AsNode()
	case lexer.Identifier:
		switch token.Value {
		case lexer.TypeKeywordInterface:
			return p.parseInterfaceDeclaration().AsNode()
		case lexer.TypeKeywordType:
			return p.parseTypeAliasDeclaration().AsNode()
		case lexer.TypeKeywordDeclare:
			modifierList := ast.NewModifierList(ast.ModifierFlagsAmbient)
			p.markStartPosition()

			p.hasPrecedingOriginalNameDirective = p.lexer.HasPrecedingOriginalNameDirective
			p.originalNameDirectiveValue = p.lexer.OriginalNameDirectiveValue
			p.expectedTypeKeyword(lexer.TypeKeywordDeclare)

			token := p.lexer.Peek()
			switch token.Type {
			case lexer.KeywordToken:
				switch token.Value {
				case lexer.KeywordLet, lexer.KeywordConst:
					return p.parseVariableStatement(modifierList).AsNode()
				case lexer.KeywordFunction:
					functionDeclaration := p.parseFunctionDeclaration(modifierList)
					if p.hasPrecedingOriginalNameDirective {
						functionDeclaration.ID.OriginalName = &p.originalNameDirectiveValue
					}
					return functionDeclaration.AsNode()
				}
			case lexer.Identifier:
				switch token.Value {
				case lexer.TypeKeywordModule:
					moduleDeclaration := p.parseModuleDeclaration()
					if p.hasPrecedingOriginalNameDirective {
						moduleDeclaration.OriginalName = &p.originalNameDirectiveValue
					}
					return moduleDeclaration.AsNode()
				}
			}
		}
	}

	if p.isIdentifier() && p.lookAhead(lexer.Colon) {
		labelledStatement := p.parseLabelledStatement()
		return labelledStatement.AsNode()
	}

	if p.isExpression() {
		p.markStartPosition()

		expression := p.parseExpression()
		p.expected(lexer.Semicolon)

		return ast.NewNode(
			ast.NewExpressionStatement(expression),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	p.errorf(
		ast.Location{
			Pos: p.startPositions.Peek(),
			End: p.getEndPosition(),
		},
		"Expected Statement got %s\n", p.lexer.Peek().Value)
	return nil
}

func (p *Parser) parseImportClause() *ast.ImportClause {
	p.markStartPosition()

	var name *ast.Node
	var namedBindings *ast.Node

	parseNamedBindings := func() *ast.Node {
		switch p.lexer.Peek().Type {
		case lexer.LeftCurlyBrace:
			return p.parseNamedImports().AsNode()
		case lexer.Star:
			return p.parseNamespaceImport().AsNode()
		}
		return nil
	}

	if p.isIdentifier() {
		name = p.parseIdentifier(false).AsNode()
		if p.optional(lexer.Comma) {
			namedBindings = parseNamedBindings()
		}
	} else {
		namedBindings = parseNamedBindings()
	}

	return ast.NewNode(
		ast.NewImportClause(name, namedBindings),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseImportDeclaration() *ast.ImportDeclaration {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordImport)

	var importClause *ast.ImportClause = nil
	if p.lexer.Peek().Type != lexer.SingleQuoteString && p.lexer.Peek().Type != lexer.DoubleQuoteString {
		importClause = p.parseImportClause()
		p.expectedKeyword(lexer.KeywordFrom)
	}

	stringLiteral := p.parseStringLiteral()
	p.optional(lexer.Semicolon)

	return ast.NewNode(
		ast.NewImportDeclaration(importClause, stringLiteral),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseNamespaceImport() *ast.NamespaceImport {
	p.markStartPosition()

	p.expected(lexer.Star)
	p.expectedKeyword(lexer.KeywordAs)
	name := p.parseIdentifier(false).AsNode()

	return ast.NewNode(
		ast.NewNamespaceImport(name),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseNamedImports() *ast.NamedImports {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	importSpecifiers := []*ast.Node{}

	importSpecifiers = append(importSpecifiers, p.parseImportSpecifier().AsNode())
	for p.optional(lexer.Comma) {
		importSpecifiers = append(importSpecifiers, p.parseImportSpecifier().AsNode())
	}

	p.optional(lexer.Comma)

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewNamedImports(importSpecifiers),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseImportSpecifier() *ast.ImportSpecifier {
	p.markStartPosition()

	var name *ast.Node = nil
	var propertyName *ast.Node = nil

	name = p.parseIdentifierName().AsNode()
	if p.optionalKeyword(lexer.KeywordAs) {
		propertyName = name
		name = p.parseIdentifierName().AsNode()
	}

	return ast.NewNode(
		ast.NewImportSpecifier(name, propertyName),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseModuleExportName() *ast.Node {
	switch {
	case p.isIdentifierName():
		return p.parseIdentifierName().AsNode()
	case p.lexer.Peek().Type == lexer.DoubleQuoteString || p.lexer.Peek().Type == lexer.SingleQuoteString:
		return p.parseStringLiteral().AsNode()
	default:
		return nil
	}
}

func (p *Parser) parseExportSpecifier() *ast.ExportSpecifier {
	p.markStartPosition()

	var name *ast.Node = nil
	var propertyName *ast.Node = nil

	name = p.parseModuleExportName()
	if p.optionalKeyword(lexer.KeywordAs) {
		propertyName = name
		name = p.parseModuleExportName()
	}

	return ast.NewNode(
		ast.NewExportSpecifier(name, propertyName),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseNamespaceExport() *ast.NamespaceExport {
	p.markStartPosition()

	p.expected(lexer.Star)

	if !p.optionalKeyword(lexer.KeywordAs) {
		p.startPositions.Pop()
		return nil
	}

	name := p.parseModuleExportName()

	return ast.NewNode(
		ast.NewNamespaceExport(name),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseNamedExports() *ast.NamedExports {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	elements := []*ast.Node{}
	if p.lexer.Peek().Type != lexer.RightCurlyBrace {
		elements = append(elements, p.parseExportSpecifier().AsNode())
		for p.optional(lexer.Comma) {
			elements = append(elements, p.parseExportSpecifier().AsNode())
		}
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewNamedExports(elements),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseExportDefaultDeclaration() *ast.ExportDefaultDeclaration {
	declOrExpr := p.parseFunctionDeclarationOrExpression()
	p.expected(lexer.Semicolon)
	return ast.NewNode(
		ast.NewExportDefaultDeclaration(declOrExpr),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseFunctionDeclarationOrExpression() *ast.Node {
	token := p.lexer.Peek()
	if token.Type == lexer.KeywordToken && token.Value == lexer.KeywordFunction {
		return p.parseFunctionDeclaration(nil).AsNode()
	}

	return p.parseExpression()
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != lexer.EOF && p.lexer.Peek().Type != lexer.Invalid && p.lexer.Peek().Type != lexer.RightCurlyBrace {
		statements = append(statements, p.parseStatement())
		p.optional(lexer.Semicolon)
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewBlockStatement(statements),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) isExpression() bool {
	token := p.lexer.Peek()
	switch token.Type {
	case lexer.LeftCurlyBrace:
		return false
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordFunction:
			return false
		default:
			return true
		}
	default:
		return true
	}
}

func (p *Parser) isPrimaryExpression() bool {
	token := p.lexer.Peek()
	switch token.Type {
	case lexer.Identifier,
		lexer.SingleQuoteString,
		lexer.DoubleQuoteString,
		lexer.Decimal,
		lexer.LeftSquareBracket,
		lexer.LeftCurlyBrace,
		lexer.LeftParenthesis,
		lexer.Slash:
		return true
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordNull, lexer.KeywordTrue, lexer.KeywordFalse, lexer.KeywordThis, lexer.KeywordFunction:
			return true
		default:
			return p.isIdentifier()
		}
	default:
		return p.isIdentifier()
	}
}

func (p *Parser) parsePrimaryExpression() *ast.Node {
	token := p.lexer.Peek()
	switch token.Type {
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordNull:
			return p.parseNullLiteral().AsNode()
		case lexer.KeywordTrue, lexer.KeywordFalse:
			return p.parseBooleanLiteral().AsNode()
		case lexer.KeywordThis:
			return p.parseThisExpression().AsNode()
		case lexer.KeywordFunction:
			return p.parseFunctionExpression().AsNode()
		default:
			if p.isIdentifier() {
				return p.parseIdentifier(false).AsNode()
			}
		}
	case lexer.Decimal:
		return p.parseDecimalLiteral().AsNode()
	case lexer.SingleQuoteString, lexer.DoubleQuoteString:
		return p.parseStringLiteral().AsNode()
	case lexer.LeftSquareBracket:
		return p.parseArrayExpression().AsNode()
	case lexer.LeftCurlyBrace:
		return p.parseObjectExpression().AsNode()
	case lexer.LeftParenthesis:
		p.expected(lexer.LeftParenthesis)
		expression := p.parseExpression()
		p.optional(lexer.Comma)
		p.expected(lexer.RightParenthesis)
		return expression
	case lexer.Slash:
		return p.parseRegularExpressionLiteral().AsNode()
	default:
		if p.isIdentifier() {
			return p.parseIdentifier(false).AsNode()
		}
	}

	p.errorf(
		ast.Location{
			Pos: p.startPositions.Peek(),
			End: p.getEndPosition(),
		},
		"Expected PrimaryExpression but got %s\n", p.lexer.Peek().Value)
	return nil
}

func (p *Parser) parseRegularExpressionLiteral() *ast.RegularExpressionLiteral {
	p.markStartPosition()
	token := p.lexer.ReScanSlashToken()
	p.lexer.Next()
	return ast.NewNode(
		ast.NewRegularExpressionLiteral(token.Value),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeNode() *ast.Node {
	p.markStartPosition()

	var typeNode *ast.Node = nil

	isUnionType := true
	types := []*ast.Node{}

	for isUnionType {
		token := p.lexer.Peek()
		switch token.Type {
		case lexer.LeftParenthesis:
			typeNode = p.parseFunctionType().AsNode()
		case lexer.LeftCurlyBrace:
			typeNode = p.parseTypeLiteral().AsNode()
		case lexer.Identifier:
			switch token.Value {
			case lexer.TypeKeywordString:
				typeNode = p.parseStringKeyword().AsNode()
			case lexer.TypeKeywordNumber:
				typeNode = p.parseNumberKeyword().AsNode()
			case lexer.TypeKeywordBoolean:
				typeNode = p.parseBooleanKeyword().AsNode()
			case lexer.TypeKeywordAny:
				typeNode = p.parseAnyKeyword().AsNode()
			default:
				typeNode = p.parseTypeReference().AsNode()
			}
		default:
			return nil
		}

		if p.optional(lexer.LeftSquareBracket) && p.optional(lexer.RightSquareBracket) {
			arrayType := ast.NewNode(
				ast.NewArrayType(typeNode),
				ast.Location{
					Pos: typeNode.Location.Pos,
					End: p.getEndPosition(),
				},
			)
			types = append(types, arrayType.AsNode())
		} else {
			types = append(types, typeNode)
		}

		if !p.optional(lexer.BitwiseOr) {
			isUnionType = false
		}
	}

	if len(types) > 1 {
		return ast.NewNode(
			ast.NewUnionType(types),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	} else {
		p.startPositions.Pop()
		return types[0]
	}
}

func (p *Parser) parseVariableStatement(modifierList *ast.ModifierList) *ast.VariableStatement {
	p.markStartPosition()

	var declarationType ast.DeclarationType = ast.DeclarationTypeNone
	switch {
	case p.optionalKeyword(lexer.KeywordLet):
		declarationType = ast.DeclarationTypeLet
	case p.optionalKeyword(lexer.KeywordConst):
		declarationType = ast.DeclarationTypeConst
	}

	variableDeclarationList := p.parseVariableDeclarationList()

	p.optional(lexer.Semicolon)

	return ast.NewNode(
		ast.NewVariableStatement(declarationType, variableDeclarationList.AsNode(), modifierList),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseVariableDeclarationList() *ast.VariableDeclarationList {
	p.markStartPosition()

	declarations := []*ast.Node{}

	variableDeclaration := p.parseVariableDeclaration()
	declarations = append(declarations, variableDeclaration.AsNode())
	if p.hasPrecedingOriginalNameDirective {
		if variableDeclaration.Name.Type == ast.NodeTypeIdentifier {
			variableDeclaration.Name.AsIdentifier().OriginalName = &p.originalNameDirectiveValue
		}
	}

	for p.optional(lexer.Comma) {
		p.hasPrecedingOriginalNameDirective = p.lexer.HasPrecedingOriginalNameDirective
		p.originalNameDirectiveValue = p.lexer.OriginalNameDirectiveValue
		variableDeclaration := p.parseVariableDeclaration()
		declarations = append(declarations, variableDeclaration.AsNode())
		if p.hasPrecedingOriginalNameDirective {
			if variableDeclaration.Name.Type == ast.NodeTypeIdentifier {
				variableDeclaration.Name.AsIdentifier().OriginalName = &p.originalNameDirectiveValue
			}
		}
	}

	return ast.NewNode(
		ast.NewVariableDeclarationList(declarations),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	p.markStartPosition()

	var name *ast.Node = nil
	switch p.lexer.Peek().Type {
	case lexer.LeftSquareBracket:
		name = p.parseArrayBindingPattern().AsNode()
	case lexer.LeftCurlyBrace:
		name = p.parseObjectBindingPattern().AsNode()
	default:
		if p.isIdentifier() {
			name = p.parseIdentifier(true).AsNode()
		}
	}

	var initializer *ast.Initializer = nil
	if p.optional(lexer.Equal) {
		assignmentExpression := p.parseAssignmentExpression()
		initializer = ast.NewNode(
			ast.NewInitializer(assignmentExpression),
			assignmentExpression.Location,
		)
	}

	return ast.NewNode(
		ast.NewVariableDeclaration(name, initializer),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseFunctionDeclaration(modifierList *ast.ModifierList) *ast.FunctionDeclaration {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordFunction)
	generator := p.optional(lexer.Star)
	identifier := p.parseIdentifier(false)
	var typeParameters *ast.TypeParametersDeclaration = nil
	if p.lexer.Peek().Type == lexer.LeftArrow {
		typeParameters = p.parseTypeParametersDeclaration()
	}

	params := p.parseParameters()

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	var body *ast.BlockStatement = nil
	if !p.optional(lexer.Semicolon) {
		body = p.parseBlockStatement()
	}

	return ast.NewNode(
		ast.NewFunctionDeclaration(identifier, typeParameters, params, body, typeAnnotation, modifierList, generator),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseCallExpressionRest(expression *ast.Node) *ast.Node {
	for {
		expression = p.parseMemberExpressionRest(expression)
		var typeParameters *ast.TypeParameterInstantiation = nil
		if p.lexer.Peek().Type == lexer.LeftArrow {
			typeParameters = p.parseTypeParameterInstantiation()
		}

		if p.optional(lexer.LeftParenthesis) {
			argumentList := []*ast.Node{}

			token := p.lexer.Peek()
			for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightParenthesis {
				pos := p.lexer.Position()
				rest := p.optional(lexer.TripleDots)
				argument := p.parseAssignmentExpression()
				if rest {
					argument = ast.NewNode(
						ast.NewSpreadElement(argument),
						ast.Location{
							Pos: uint(pos),
							End: argument.Location.End,
						},
					).AsNode()
				}
				argumentList = append(argumentList, argument)
				token = p.lexer.Peek()
				if token.Type == lexer.Comma {
					token = p.lexer.Next()
				}
			}

			p.expected(lexer.RightParenthesis)
			expression = ast.NewNode(
				ast.NewCallExpression(expression, typeParameters, argumentList),
				ast.Location{
					Pos: expression.Location.Pos,
					End: p.getEndPosition(),
				},
			).AsNode()
			continue
		}

		break
	}
	return expression
}

func (p *Parser) parseLeftHandSideExpression() *ast.Node {
	return p.parseCallExpressionRest(p.parseMemberExpression())
}

func (p *Parser) parseMemberExpression() *ast.Node {
	if p.isPrimaryExpression() {
		memberExpression := p.parsePrimaryExpression()
		return p.parseMemberExpressionRest(memberExpression)
	}

	p.errorf(
		ast.Location{
			Pos: p.startPositions.Peek(),
			End: p.getEndPosition(),
		},
		"Expected MemberExpression but got %s\n", p.lexer.Peek().Value)
	return nil
}

func (p *Parser) parseMemberExpressionRest(memberExpression *ast.Node) *ast.Node {
	for {
		if p.optional(lexer.LeftSquareBracket) {
			computed := true
			property := p.parseExpression()
			memberExpression = ast.NewNode(
				ast.NewMemberExpression(memberExpression, property, computed),
				ast.Location{
					Pos: memberExpression.Location.Pos,
					End: p.getEndPosition(),
				},
			).AsNode()
			p.expected(lexer.RightSquareBracket)
		} else if p.optional(lexer.Dot) {
			computed := false
			property := p.parseIdentifier(false).AsNode()
			memberExpression = ast.NewNode(
				ast.NewMemberExpression(memberExpression, property, computed),
				ast.Location{
					Pos: memberExpression.Location.Pos,
					End: p.getEndPosition(),
				},
			).AsNode()
		} else {
			break
		}
	}
	return memberExpression
}

func (p *Parser) parseUpdateExpression() *ast.Node {
	p.markStartPosition()

	isUpdateExpression := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.DoublePlus) || p.optional(lexer.DoubleMinus))
	}

	if operator, ok := isUpdateExpression(); ok {
		return ast.NewNode(
			ast.NewUpdateExpression(operator, p.parseUnaryExpression(), true),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	leftHandSideExpression := p.parseLeftHandSideExpression()

	if operator, ok := isUpdateExpression(); ok {
		return ast.NewNode(
			ast.NewUpdateExpression(operator, leftHandSideExpression, false),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	p.startPositions.Pop()

	return leftHandSideExpression
}

func (p *Parser) parseUnaryExpression() *ast.Node {
	p.markStartPosition()

	isPrefixUnaryExpression := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.Plus) || p.optional(lexer.Minus))
	}

	if operator, ok := isPrefixUnaryExpression(); ok {
		return ast.NewNode(
			ast.NewPrefixUnaryExpression(operator, p.parseUnaryExpression()),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	p.startPositions.Pop()

	return p.parseUpdateExpression()
}

func (p *Parser) parseExponentiationExpression() *ast.Node {
	isUpdateExpression := func(tokenType lexer.TokenType) bool {
		return tokenType != lexer.Plus && tokenType != lexer.Minus
	}
	token := p.lexer.Peek()
	left := p.parseUnaryExpression()
	if !isUpdateExpression(token.Type) {
		return left
	}

	if p.optional(lexer.DoubleStar) {
		right := p.parseExponentiationExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression("**", left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseMultiplicativeExpression() *ast.Node {
	isMultiplicativeToken := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.Slash) || p.optional(lexer.Percent) || p.optional(lexer.Star))
	}
	left := p.parseExponentiationExpression()
	for operator, ok := isMultiplicativeToken(); ok; operator, ok = isMultiplicativeToken() {
		right := p.parseExponentiationExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression(operator, left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseAdditiveExpression() *ast.Node {
	isAdditiveToken := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.Plus) || p.optional(lexer.Minus))
	}

	left := p.parseMultiplicativeExpression()
	for operator, ok := isAdditiveToken(); ok; operator, ok = isAdditiveToken() {
		right := p.parseMultiplicativeExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression(operator, left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseShiftExpression() *ast.Node {
	isShiftToken := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.DoubleLeftArrow) ||
			p.optional(lexer.DoubleRightArrow) ||
			p.optional(lexer.TripleRightArrow))
	}

	left := p.parseAdditiveExpression()
	for operator, ok := isShiftToken(); ok; operator, ok = isShiftToken() {
		right := p.parseAdditiveExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression(operator, left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseRelationalExpression() *ast.Node {
	isRelationalToken := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.LeftArrow) ||
			p.optional(lexer.RightArrow) ||
			p.optional(lexer.LeftArrowEqual) ||
			p.optional(lexer.RightArrowEqual))
	}

	left := p.parseShiftExpression()
	for operator, ok := isRelationalToken(); ok; operator, ok = isRelationalToken() {
		right := p.parseShiftExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression(operator, left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	return left
}

func (p *Parser) parseEqualityExpression() *ast.Node {
	isEqualityToken := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(lexer.EqualEqual) ||
			p.optional(lexer.EqualEqualEqual) ||
			p.optional(lexer.NotEqual) ||
			p.optional(lexer.NotEqualEqual))
	}

	left := p.parseRelationalExpression()
	for operator, ok := isEqualityToken(); ok; operator, ok = isEqualityToken() {
		right := p.parseRelationalExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression(operator, left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseBitwiseANDExpression() *ast.Node {
	left := p.parseEqualityExpression()
	for p.optional(lexer.BitwiseAnd) {
		right := p.parseEqualityExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression("&", left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseBitwiseXORExpression() *ast.Node {
	left := p.parseBitwiseANDExpression()
	for p.optional(lexer.BitwiseXor) {
		right := p.parseBitwiseANDExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression("^", left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseBitwiseORExpression() *ast.Node {
	left := p.parseBitwiseXORExpression()
	for p.optional(lexer.BitwiseOr) {
		right := p.parseBitwiseXORExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression("|", left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseLogicalANDExpression() *ast.Node {
	return p.parseBitwiseORExpression()
}

func (p *Parser) parseLogicalORExpression() *ast.Node {
	return p.parseLogicalANDExpression()
}

func (p *Parser) parseShortCircuitExpression() *ast.Node {
	return p.parseLogicalORExpression()
}

func (p *Parser) parseConditionalExpression() *ast.Node {
	p.markStartPosition()
	shortCircuitExpression := p.parseShortCircuitExpression()
	if p.optional(lexer.Question) {
		whenTrue := p.parseAssignmentExpression()
		p.expected(lexer.Colon)
		whenFalse := p.parseAssignmentExpression()

		return ast.NewNode(
			ast.NewConditionalExpression(shortCircuitExpression, whenTrue, whenFalse),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	p.startPositions.Pop()
	return shortCircuitExpression
}

func (p *Parser) tryParseParenthesizedArrowFunction() *ast.ArrowFunction {
	state := p.mark()
	p.expected(lexer.LeftParenthesis)
	arrowFunctionExpression := p.tryParseArrowFunction()
	if arrowFunctionExpression == nil {
		p.rewind(state)
		return nil
	}
	p.expected(lexer.RightParenthesis)
	return arrowFunctionExpression
}

func (p *Parser) tryParseArrowFunction() *ast.ArrowFunction {
	state := p.mark()
	p.markStartPosition()

	params := p.parseArrowParametes()
	if !p.expected(lexer.EqualRightArrow) {
		p.rewind(state)
		return nil
	}

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	var body *ast.Node = nil
	if p.lexer.Peek().Type == lexer.LeftCurlyBrace {
		body = p.parseBlockStatement().AsNode()
	} else {
		body = p.parseAssignmentExpression()
	}

	return ast.NewNode(
		ast.NewArrowFunction(nil, params, body, typeAnnotation),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseArrowParametes() []*ast.Node {
	if p.isIdentifier() {
		return []*ast.Node{p.parseIdentifier(false).AsNode()}
	}

	return p.parseParameters()
}

func (p *Parser) parseParameters() []*ast.Node {
	p.expected(lexer.LeftParenthesis)
	params := []*ast.Node{}
	param := p.parseParameter()
	if param != nil {
		params = append(params, param)
	}
	for p.optional(lexer.Comma) {
		param := p.parseParameter()
		if param != nil {
			params = append(params, param)
		}
	}
	p.expected(lexer.RightParenthesis)
	return params
}

func (p *Parser) parseObjectBindingElement() *ast.BindingElement {
	p.markStartPosition()

	rest := p.optional(lexer.TripleDots)
	var propertyName *ast.Node = nil
	name := p.parsePropertyName()
	if p.optional(lexer.Colon) {
		propertyName = name
		name = p.parseIdentifierOrPattern()
	}

	var initializer *ast.Initializer = nil
	if p.lexer.Peek().Type == lexer.Equal {
		p.expected(lexer.Equal)
		assignmentExpression := p.parseAssignmentExpression()
		initializer = ast.NewNode(
			ast.NewInitializer(assignmentExpression),
			assignmentExpression.Location,
		)
	}

	return ast.NewNode(
		ast.NewBindingElement(name, propertyName, rest, nil, initializer),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseIdentifierOrPattern() *ast.Node {
	switch p.lexer.Peek().Type {
	case lexer.LeftCurlyBrace:
		return p.parseObjectBindingPattern().AsNode()
	case lexer.LeftSquareBracket:
		return p.parseArrayBindingPattern().AsNode()
	default:
		if p.isIdentifier() {
			return p.parseIdentifier(false).AsNode()
		}
		return nil
	}
}

func (p *Parser) parseParameter() *ast.Node {
	p.markStartPosition()

	rest := p.optional(lexer.TripleDots)
	element := p.parseIdentifierOrPattern()
	if element == nil {
		p.startPositions.Pop()
		return nil
	}

	isOptionalParameter := p.optional(lexer.Question)

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	var initializer *ast.Initializer = nil
	if p.lexer.Peek().Type == lexer.Equal {
		p.expected(lexer.Equal)
		assignmentExpression := p.parseAssignmentExpression()
		initializer = ast.NewNode(
			ast.NewInitializer(assignmentExpression),
			assignmentExpression.Location,
		)
	}

	return ast.NewNode(
		ast.NewParameter(element.AsNode(), rest, isOptionalParameter, typeAnnotation, initializer),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	).AsNode()
}

func (p *Parser) parsePropertyName() *ast.Node {
	var name *ast.Node = nil

	switch p.lexer.Peek().Type {
	case lexer.DoubleQuoteString, lexer.SingleQuoteString:
		name = p.parseStringLiteral().AsNode()
	case lexer.Decimal:
		name = p.parseDecimalLiteral().AsNode()
	case lexer.LeftSquareBracket:
		p.expected(lexer.LeftSquareBracket)
		assignmentExpression := p.parseAssignmentExpression()
		name = ast.NewNode(
			ast.NewComputedProperty(assignmentExpression.AsNode()),
			assignmentExpression.Location,
		).AsNode()
		p.expected(lexer.RightSquareBracket)
	default:
		if p.isIdentifier() {
			name = p.parseIdentifier(false).AsNode()
		}
	}

	return name
}

func (p *Parser) parseObjectBindingPattern() *ast.ObjectBindingPattern {
	p.markStartPosition()
	p.expected(lexer.LeftCurlyBrace)

	elements := []*ast.Node{}
	for {
		elements = append(elements, p.parseObjectBindingElement().AsNode())
		if !p.optional(lexer.Comma) {
			break
		}
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewObjectBindingPattern(elements),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseArrayBindingElement() *ast.BindingElement {
	p.markStartPosition()

	rest := p.optional(lexer.TripleDots)
	name := p.parseIdentifierOrPattern()

	var initializer *ast.Initializer = nil
	if p.lexer.Peek().Type == lexer.Equal {
		p.expected(lexer.Equal)
		assignmentExpression := p.parseAssignmentExpression()
		initializer = ast.NewNode(
			ast.NewInitializer(assignmentExpression),
			assignmentExpression.Location,
		)
	}

	return ast.NewNode(
		ast.NewBindingElement(name, nil, rest, nil, initializer),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseArrayBindingPattern() *ast.ArrayBindingPattern {
	p.markStartPosition()
	elements := []*ast.Node{}
	p.expected(lexer.LeftSquareBracket)
	for {
		p.optional(lexer.Comma)
		bindingElement := p.parseArrayBindingElement()
		if bindingElement != nil {
			elements = append(elements, bindingElement.AsNode())
		}
		if !p.optional(lexer.Comma) {
			break
		}
	}
	p.expected(lexer.RightSquareBracket)

	return ast.NewNode(
		ast.NewArrayBindingPattern(elements),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseAssignmentExpression() *ast.Node {
	if p.lexer.Peek().Type == lexer.LeftParenthesis {
		parenthesizedArrowFunction := p.tryParseParenthesizedArrowFunction()
		if parenthesizedArrowFunction != nil {
			return parenthesizedArrowFunction.AsNode()
		}
	}

	arrowFunctionExpression := p.tryParseArrowFunction()
	if arrowFunctionExpression != nil {
		return arrowFunctionExpression.AsNode()
	}

	node := p.parseConditionalExpression()
	if node == nil {
		return nil
	}

	switch node.Type {
	case
		ast.NodeTypeIdentifier,
		ast.NodeTypeStringLiteral,
		ast.NodeTypeDecimalLiteral,
		ast.NodeTypeNullLiteral,
		ast.NodeTypeBooleanLiteral,
		ast.NodeTypeArrayExpression,
		ast.NodeTypeObjectExpression,
		ast.NodeTypeMemberExpression,
		ast.NodeTypeCallExpression:
		{
			token := p.lexer.Peek()
			switch token.Type {
			case
				lexer.Equal,
				lexer.StarEqual,
				lexer.SlashEqual,
				lexer.PercentEqual,
				lexer.PlusEqual,
				lexer.MinusEqual,
				lexer.BitwiseAndEqual,
				lexer.BitwiseXorEqual,
				lexer.BitwiseOrEqual,
				lexer.DoubleLeftArrowEqual,
				lexer.DoubleRightArrowEqual,
				lexer.DoubleStarEqual:
				{
					p.lexer.Next()
					return ast.NewNode(
						ast.NewAssignmentExpression(token.Value, node, p.parseAssignmentExpression()),
						ast.Location{
							Pos: node.Location.Pos,
							End: p.getEndPosition(),
						},
					).AsNode()
				}
			}
		}
	}

	return node
}

func (p *Parser) parseExpression() *ast.Node {
	left := p.parseAssignmentExpression()
	// Comma Operator: https://tc39.es/ecma262/#sec-comma-operator
	for p.optional(lexer.Comma) {
		right := p.parseAssignmentExpression()
		left = ast.NewNode(
			ast.NewBinaryExpression(",", left, right),
			ast.Location{
				Pos: left.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}
	return left
}

func (p *Parser) parseInterfaceDeclaration() *ast.InterfaceDeclaration {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordInterface)

	identifier := p.parseIdentifier(false)
	var typeParameters *ast.TypeParametersDeclaration = nil
	if p.lexer.Peek().Type == lexer.LeftArrow {
		typeParameters = p.parseTypeParametersDeclaration()
	}

	extends := []*ast.Node{}
	if p.optionalKeyword(lexer.KeywordExtends) {
		for {
			extends = append(extends, p.parseIdentifier(false).AsNode())
			if !p.optional(lexer.Comma) {
				break
			}
		}
	}

	body := p.parseInterfaceBody()
	p.optional(lexer.Semicolon)

	return ast.NewNode(
		ast.NewInterfaceDeclaration(identifier, typeParameters, body, extends),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parsePropertySignature() *ast.PropertySignature {
	p.markStartPosition()

	var key *ast.Node = nil
	hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
	originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue

	var modifierList *ast.ModifierList = nil
	if p.optionalTypeKeyword(lexer.TypeKeywordReadOnly) {
		modifierList = ast.NewModifierList(ast.ModifierFlagsReadonly)
	}

	switch p.lexer.Peek().Type {
	case lexer.DoubleQuoteString, lexer.SingleQuoteString:
		key = p.parseStringLiteral().AsNode()
	case lexer.Decimal:
		key = p.parseDecimalLiteral().AsNode()
	default:
		if p.isIdentifier() {
			identifier := p.parseIdentifier(false)
			if hasPrecedingOriginalNameDirective {
				identifier.OriginalName = &originalNameDirectiveValue
			}

			key = identifier.AsNode()
		}
	}

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	return ast.NewNode(
		ast.NewPropertySignature(key, typeAnnotation, modifierList),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseInterfaceBody() *ast.InterfaceBody {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	body := []*ast.Node{}

	for {
		token := p.lexer.Peek()
		if token.Type == lexer.EOF || token.Type == lexer.Invalid || token.Type == lexer.RightCurlyBrace {
			break
		}

		switch token.Type {
		case lexer.Identifier, lexer.DoubleQuoteString, lexer.SingleQuoteString, lexer.Decimal:
			body = append(body, p.parsePropertySignature().AsNode())
		case lexer.LeftSquareBracket:
			body = append(body, p.parseIndexSignatureDeclaration().AsNode())
		}

		if p.optional(lexer.Semicolon) || p.optional(lexer.Comma) {
			continue
		}
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewInterfaceBody(body),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	p.markStartPosition()
	return ast.NewNode(
		ast.NewTypeAnnotation(p.parseTypeNode()),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordReturn)
	argument := p.parseExpression()
	p.expected(lexer.Semicolon)

	return ast.NewNode(
		ast.NewReturnStatement(argument),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseFunctionType() *ast.FunctionType {
	p.markStartPosition()

	params := p.parseParameters()
	p.expected(lexer.EqualRightArrow)

	typeAnnotation := p.parseTypeAnnotation()

	return ast.NewNode(
		ast.NewFunctionType(params, typeAnnotation),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeAliasDeclaration() *ast.TypeAliasDeclaration {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordType)
	identifier := p.parseIdentifier(false)
	var typeParameters *ast.TypeParametersDeclaration = nil
	if p.lexer.Peek().Type == lexer.LeftArrow {
		typeParameters = p.parseTypeParametersDeclaration()
	}
	p.expected(lexer.Equal)
	typeAnnotation := p.parseTypeAnnotation()

	return ast.NewNode(
		ast.NewTypeAliasDeclaration(identifier, typeParameters, typeAnnotation),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeLiteral() *ast.TypeLiteral {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	members := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightCurlyBrace {
		var key *ast.Node
		var modifierList *ast.ModifierList = nil
		if p.optionalTypeKeyword(lexer.TypeKeywordReadOnly) {
			modifierList = ast.NewModifierList(ast.ModifierFlagsReadonly)
		}
		switch token.Type {
		case lexer.DoubleQuoteString, lexer.SingleQuoteString:
			key = p.parseStringLiteral().AsNode()
		case lexer.Decimal:
			key = p.parseDecimalLiteral().AsNode()
		default:
			if p.isIdentifier() {
				key = p.parseIdentifier(false).AsNode()
				break
			}
			p.errorf(
				ast.Location{
					Pos: p.startPositions.Pop(),
					End: p.getEndPosition(),
				},
				"Expected a valid key but got %s\n", token.Value)
			return nil
		}

		if p.optional(lexer.Colon) {
			members = append(members,
				ast.NewNode(
					ast.NewPropertySignature(key, p.parseTypeAnnotation(), modifierList),
					ast.Location{
						Pos: key.Location.Pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)

			p.optional(lexer.Comma)
			token = p.lexer.Peek()
		} else {
			if token.Type != lexer.Comma && token.Type != lexer.RightCurlyBrace {
				p.errorf(
					ast.Location{
						Pos: p.startPositions.Pop(),
						End: p.getEndPosition(),
					},
					"A Expected '{' but got %s\n", token.Value)
				return nil
			}
			p.optional(lexer.Comma)
			if key.Type == ast.NodeTypeIdentifier {
				members = append(members,
					ast.NewNode(
						ast.NewPropertySignature(key, p.parseTypeAnnotation(), modifierList),
						ast.Location{
							Pos: key.Location.Pos,
							End: p.getEndPosition(),
						},
					).AsNode(),
				)
			} else {
				p.errorf(
					ast.Location{
						Pos: p.startPositions.Pop(),
						End: p.getEndPosition(),
					},
					"Expected Identifier but got %s\n", token.Value)
				return nil
			}
		}

		token = p.lexer.Peek()
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewTypeLiteral(members),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseBreakableStatement() *ast.Node {
	token := p.lexer.Peek()
	switch token.Value {
	case lexer.KeywordFor:
		return p.parseIterationStatement()
	}

	return nil
}

func (p *Parser) parseIterationStatement() *ast.Node {
	token := p.lexer.Peek()
	switch token.Value {
	case lexer.KeywordFor:
		return p.parseForStatement().AsNode()
	}

	return nil
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordFor)
	p.expected(lexer.LeftParenthesis)

	var init *ast.Node = nil
	token := p.lexer.Peek()
	if token.Type == lexer.KeywordToken && (token.Value == lexer.KeywordLet || token.Value == lexer.KeywordConst) {
		init = p.parseVariableStatement(nil).AsNode()
	} else {
		init = p.parseExpression()
		p.expected(lexer.Semicolon)
	}

	test := p.parseExpression()
	p.expected(lexer.Semicolon)
	update := p.parseExpression()
	p.expected(lexer.RightParenthesis)

	body := p.parseStatement()

	return ast.NewNode(
		ast.NewForStatement(init, test, update, body),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) isIdentifierName() bool {
	return p.lexer.Peek().Type == lexer.Identifier || p.lexer.Peek().Type == lexer.KeywordToken
}

func (p *Parser) parseIdentifierName() *ast.Identifier {
	p.markStartPosition()

	token := p.lexer.Peek()
	identifierName := token.Value

	if !p.isIdentifierName() {
		pos := p.startPositions.Pop()
		return ast.NewNode(
			ast.NewIdentifier(identifierName, nil),
			ast.Location{
				Pos: pos,
				End: pos,
			},
		)
	}

	p.lexer.Next()

	return ast.NewNode(
		ast.NewIdentifier(identifierName, nil),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) isIdentifier() bool {
	return p.lexer.Peek().Type == lexer.Identifier ||
		(p.lexer.Peek().Type == lexer.KeywordToken && !lexer.IsReservedKeyword(p.lexer.Peek()))
}

func (p *Parser) parseIdentifier(doParseTypeAnnotation bool) *ast.Identifier {
	p.markStartPosition()

	identifierName := p.lexer.Peek().Value
	if !p.isIdentifier() {
		pos := uint(p.lexer.BeforeWhitespacePosition())
		p.startPositions.Pop()
		return ast.NewNode(
			ast.NewIdentifier(identifierName, nil),
			ast.Location{
				Pos: pos,
				End: pos,
			},
		)
	}
	p.lexer.Next()

	var typeAnnotation *ast.TypeAnnotation = nil
	if doParseTypeAnnotation && p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	return ast.NewNode(
		ast.NewIdentifier(identifierName, typeAnnotation),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseStringLiteral() *ast.StringLiteral {
	p.markStartPosition()

	stringLiteralValue := p.lexer.Peek().Value

	if !p.optional(lexer.DoubleQuoteString) && !p.optional(lexer.SingleQuoteString) {
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
			"Expected string but got '%s'\n", p.lexer.Peek().Value)
		return nil
	}

	return ast.NewNode(
		ast.NewStringLiteral(stringLiteralValue),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseNullLiteral() *ast.NullLiteral {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordNull)

	return ast.NewNode(
		ast.NewNullLiteral(),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseDecimalLiteral() *ast.DecimalLiteral {
	p.markStartPosition()

	value := p.lexer.Peek().Value
	p.expected(lexer.Decimal)

	return ast.NewNode(
		ast.NewDecimalLiteral(value),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseBooleanLiteral() *ast.BooleanLiteral {
	p.markStartPosition()

	value := p.lexer.Peek().Value
	if !p.optionalKeyword(lexer.KeywordTrue) && !p.optionalKeyword(lexer.KeywordFalse) {
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
			"Expected boolean but got '%s'\n", p.lexer.Peek().Value)
		return nil
	}

	return ast.NewNode(
		ast.NewBooleanLiteral(value == lexer.KeywordTrue),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseExportClause() *ast.Node {
	switch p.lexer.Peek().Type {
	case lexer.LeftCurlyBrace:
		return p.parseNamedExports().AsNode()
	case lexer.Star:
		namespaceExport := p.parseNamespaceExport()
		if namespaceExport != nil {
			return namespaceExport.AsNode()
		}
		return nil
	default:
		return nil
	}
}

func (p *Parser) parseExportDeclarationOrExportAssignment() *ast.Node {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordExport)

	if p.optionalKeyword(lexer.KeywordDefault) {
		assignmentExpression := p.parseAssignmentExpression()
		p.optional(lexer.Semicolon)
		return ast.NewNode(
			ast.NewExportAssignment(assignmentExpression),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	switch p.lexer.Peek().Type {
	case lexer.KeywordToken:
		switch p.lexer.Peek().Value {
		case lexer.KeywordLet, lexer.KeywordConst:
			p.startPositions.Pop()
			return p.parseVariableStatement(&ast.ModifierList{ModifierFlags: ast.ModifierFlagsExport}).AsNode()
		case lexer.KeywordFunction:
			p.startPositions.Pop()
			return p.parseFunctionDeclaration(&ast.ModifierList{ModifierFlags: ast.ModifierFlagsExport}).AsNode()
		}
	}

	exportClause := p.parseExportClause()
	var moduleSpecifier *ast.StringLiteral = nil
	if p.optionalKeyword(lexer.KeywordFrom) {
		moduleSpecifier = p.parseStringLiteral()
	}

	p.optional(lexer.Semicolon)

	return ast.NewNode(
		ast.NewExportDeclaration(exportClause, moduleSpecifier),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	).AsNode()
}

func (p *Parser) parseDirective() *ast.Directive {
	p.markStartPosition()

	stringLiteralValue := p.lexer.Peek().Value

	if !p.optional(lexer.DoubleQuoteString) && !p.optional(lexer.SingleQuoteString) {
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
			"Expected string but got '%s'\n", p.lexer.Peek().Value)
		return nil
	}

	directiveLiteral := ast.NewNode(
		ast.NewDirectiveLiteral(stringLiteralValue),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)

	return ast.NewNode(
		ast.NewDirective(directiveLiteral),
		directiveLiteral.Location,
	)
}

func (p *Parser) parseDirectives() []*ast.Directive {
	directives := []*ast.Directive{}

	token := p.lexer.Peek()
	for token.Type == lexer.SingleQuoteString || token.Type == lexer.DoubleQuoteString {
		directives = append(directives, p.parseDirective())
		p.expected(lexer.Semicolon)
		token = p.lexer.Peek()
	}

	return directives
}

func (p *Parser) parseArrayExpression() *ast.ArrayExpression {
	p.markStartPosition()

	p.expected(lexer.LeftSquareBracket)

	elements := []*ast.Node{}
	token := p.lexer.Peek()

	for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightSquareBracket {
		for p.optional(lexer.Comma) {
		}

		if p.lexer.Peek().Type == lexer.RightSquareBracket {
			break
		}

		pos := uint(p.lexer.Position())
		if p.optional(lexer.TripleDots) {
			elements = append(elements,
				ast.NewNode(
					ast.NewSpreadElement(p.parseAssignmentExpression()),
					ast.Location{
						Pos: pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)
		} else {
			elements = append(elements, p.parseAssignmentExpression())
		}

		token = p.lexer.Peek()
	}

	p.expected(lexer.RightSquareBracket)

	return ast.NewNode(
		ast.NewArrayExpression(elements),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}
func (p *Parser) parseObjectPropertyOrMethod() *ast.Node {
	p.markStartPosition()

	var key *ast.Node = nil
	var value *ast.Node = nil
	switch p.lexer.Peek().Type {
	case lexer.DoubleQuoteString, lexer.SingleQuoteString:
		key = p.parseStringLiteral().AsNode()
	case lexer.Decimal:
		key = p.parseDecimalLiteral().AsNode()
	case lexer.LeftSquareBracket:
		p.expected(lexer.LeftSquareBracket)
		assignmentExpression := p.parseAssignmentExpression()
		key = ast.NewNode(
			ast.NewComputedProperty(assignmentExpression.AsNode()),
			assignmentExpression.Location,
		).AsNode()
		p.expected(lexer.RightSquareBracket)
	default:
		generator := p.optional(lexer.Star)
		if !p.isIdentifierName() {
			break
		}

		identifier := p.parseIdentifierName()
		if p.lexer.Peek().Type == lexer.LeftParenthesis || p.lexer.Peek().Type == lexer.LeftArrow {
			var typeParameters *ast.TypeParametersDeclaration = nil
			if p.lexer.Peek().Type == lexer.LeftArrow {
				typeParameters = p.parseTypeParametersDeclaration()
			}

			params := p.parseParameters()

			var typeAnnotation *ast.TypeAnnotation = nil
			if p.optional(lexer.Colon) {
				typeAnnotation = p.parseTypeAnnotation()
			}

			var body *ast.BlockStatement = nil
			if !p.optional(lexer.Semicolon) {
				body = p.parseBlockStatement()
			}

			return ast.NewNode(
				ast.NewObjectMethod(identifier, typeParameters, params, body, typeAnnotation, generator),
				ast.Location{
					Pos: p.startPositions.Pop(),
					End: p.getEndPosition(),
				},
			).AsNode()
		}

		if p.optional(lexer.Equal) {
			assignmentExpression := p.parseAssignmentExpression()
			initializer := ast.NewNode(
				ast.NewInitializer(assignmentExpression),
				assignmentExpression.Location,
			)
			return ast.NewNode(
				ast.NewShorthandPropertyAssignment(identifier.AsNode(), initializer),
				ast.Location{
					Pos: p.startPositions.Pop(),
					End: p.getEndPosition(),
				},
			).AsNode()
		}

		key = identifier.AsNode()
	}

	if p.optional(lexer.Colon) {
		value = p.parseAssignmentExpression()
	} else {
		if p.lexer.Peek().Type == lexer.Comma {
			value = key
		}
	}

	return ast.NewNode(
		ast.NewObjectProperty(key, value),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	).AsNode()
}

func (p *Parser) parseObjectExpression() *ast.ObjectExpression {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	properties := []*ast.Node{}

	for {
		doBreak := false
		token := p.lexer.Peek()
		switch token.Type {
		case lexer.EOF, lexer.Invalid, lexer.RightCurlyBrace:
			doBreak = true
		case lexer.TripleDots:
			p.optional(lexer.TripleDots)
			properties = append(properties,
				ast.NewNode(
					ast.NewSpreadElement(p.parseAssignmentExpression()),
					ast.Location{
						Pos: uint(p.lexer.Position()),
						End: p.getEndPosition(),
					},
				).AsNode(),
			)
		default:
			properties = append(properties, p.parseObjectPropertyOrMethod())
		}

		if doBreak {
			break
		}

		token = p.lexer.Peek()
		if token.Type != lexer.Comma && token.Type != lexer.RightCurlyBrace {
			p.errorf(
				ast.Location{
					Pos: p.startPositions.Pop(),
					End: p.getEndPosition(),
				},
				"Expected '{' but got %s\n", token.Value)
			return nil
		}

		p.optional(lexer.Comma)
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewObjectExpression(properties),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseStringKeyword() *ast.StringKeyword {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordString)

	return ast.NewNode(
		ast.NewStringKeyword(),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseNumberKeyword() *ast.NumberKeyword {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordNumber)

	return ast.NewNode(
		ast.NewNumberKeyword(),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseBooleanKeyword() *ast.BooleanKeyword {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordBoolean)

	return ast.NewNode(
		ast.NewBooleanKeyword(),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseAnyKeyword() *ast.AnyKeyword {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordAny)

	return ast.NewNode(
		ast.NewAnyKeyword(),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeReference() *ast.TypeReferenceNode {
	p.markStartPosition()

	identifier := p.parseIdentifier(false)

	var typeParameters *ast.TypeParameterInstantiation = nil
	if p.lexer.Peek().Type == lexer.LeftArrow {
		typeParameters = p.parseTypeParameterInstantiation()
	}

	return ast.NewNode(
		ast.NewTypeReferenceNode(identifier, typeParameters),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseModuleDeclaration() *ast.ModuleDeclaration {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordModule)
	id := p.parseStringLiteral()
	body := p.parseModuleBlock()

	return ast.NewNode(
		ast.NewModuleDeclaration(id, body),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseModuleBlock() *ast.ModuleBlock {
	p.markStartPosition()

	body := []*ast.Node{}

	p.expected(lexer.LeftCurlyBrace)
	token := p.lexer.Peek()
	for token.Type != lexer.RightCurlyBrace {
		switch token.Type {
		case lexer.KeywordToken:
			switch token.Value {
			case lexer.KeywordLet, lexer.KeywordConst:
				body = append(body, p.parseVariableStatement(nil).AsNode())
			case lexer.KeywordFunction:
				body = append(body, p.parseFunctionDeclaration(nil).AsNode())
			}
		case lexer.Identifier:
			switch token.Value {
			case lexer.TypeKeywordInterface:
				body = append(body, p.parseInterfaceDeclaration().AsNode())
			}
		}
		token = p.lexer.Peek()
	}

	p.expected(lexer.RightCurlyBrace)

	return ast.NewNode(
		ast.NewModuleBlock(body),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeParametersDeclaration() *ast.TypeParametersDeclaration {
	p.markStartPosition()

	params := []*ast.TypeParameter{}

	p.expected(lexer.LeftArrow)

	identifier := p.parseIdentifier(false)
	params = append(params, ast.NewNode(ast.NewTypeParameter(identifier.Name), identifier.Location))

	for p.optional(lexer.Comma) {
		identifier := p.parseIdentifier(false)
		params = append(params, ast.NewNode(ast.NewTypeParameter(identifier.Name), identifier.Location))
	}

	p.expected(lexer.RightArrow)

	return ast.NewNode(
		ast.NewTypeParametersDeclaration(params),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeParameterInstantiation() *ast.TypeParameterInstantiation {
	state := p.mark()
	p.markStartPosition()

	params := []*ast.Node{}
	p.expected(lexer.LeftArrow)
	params = append(params, p.parseTypeNode())
	for p.optional(lexer.Comma) {
		params = append(params, p.parseTypeNode())
	}

	if !p.expected(lexer.RightArrow) {
		p.rewind(state)
		return nil
	}

	return ast.NewNode(
		ast.NewTypeParameterInstantiation(params),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseThisExpression() *ast.ThisEpxression {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordThis)

	return ast.NewNode(
		ast.NewThisEpxression(),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseFunctionExpression() *ast.FunctionExpression {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordFunction)

	var identifier *ast.Identifier = nil
	if p.isIdentifier() {
		identifier = p.parseIdentifier(false)
	}

	var typeParameters *ast.TypeParametersDeclaration = nil
	if p.lexer.Peek().Type == lexer.LeftArrow {
		typeParameters = p.parseTypeParametersDeclaration()
	}

	params := p.parseParameters()

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	body := p.parseBlockStatement()

	return ast.NewNode(
		ast.NewFunctionExpression(identifier, typeParameters, params, body, typeAnnotation),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseIndexSignatureDeclaration() *ast.IndexSignatureDeclaration {
	p.markStartPosition()

	p.expected(lexer.LeftSquareBracket)
	index := p.parseIdentifier(true)
	p.expected(lexer.RightSquareBracket)
	p.expected(lexer.Colon)
	typeNode := p.parseTypeNode()

	return ast.NewNode(
		ast.NewIndexSignatureDeclaration(index, typeNode),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) expected(tokenType lexer.TokenType) bool {
	token := p.lexer.Peek()
	if token.Type != tokenType {
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Peek(),
				End: p.getEndPosition(),
			},
			"expected '%s' but got '%s'\n", tokenType.String(), token.Type.String())
		return false
	}
	p.lexer.Next()
	return true
}

func (p *Parser) expectedKeyword(keyword lexer.Keyword) {
	token := p.lexer.Peek()
	if token.Type != lexer.KeywordToken && token.Value != keyword {
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Peek(),
				End: p.getEndPosition(),
			},
			"expected '%s' keyword but got '%s'\n", keyword, token.Value)
	}
	p.lexer.Next()
}

func (p *Parser) expectedTypeKeyword(typeKeyword lexer.TypeKeyword) {
	token := p.lexer.Peek()
	if token.Type != lexer.Identifier && token.Value != typeKeyword {
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Peek(),
				End: p.getEndPosition(),
			},
			"expected '%s' type keyword but got '%s'\n", typeKeyword, token.Value)
	}
	p.lexer.Next()
}

func (p *Parser) optional(tokenType lexer.TokenType) bool {
	token := p.lexer.Peek()
	if token.Type == tokenType {
		p.lexer.Next()
		return true
	}
	return false
}

func (p *Parser) optionalKeyword(keyword lexer.Keyword) bool {
	token := p.lexer.Peek()
	if token.Type == lexer.KeywordToken && token.Value == keyword {
		p.lexer.Next()
		return true
	}
	return false
}

func (p *Parser) optionalTypeKeyword(typeKeyword lexer.TypeKeyword) bool {
	token := p.lexer.Peek()
	if token.Type == lexer.Identifier && token.Value == typeKeyword {
		p.lexer.Next()
		return true
	}
	return false
}

func (p *Parser) lookAhead(tokenType lexer.TokenType) bool {
	return p.lexer.LookAhead().Type == tokenType
}

func (p *Parser) lookAheadKeyword(keyword lexer.Keyword) bool {
	token := p.lexer.LookAhead()
	return token.Type == lexer.KeywordToken && token.Value == keyword
}

func (p *Parser) lookAheadTypeKeyword(typeKeyword lexer.TypeKeyword) bool {
	token := p.lexer.LookAhead()
	return token.Type == lexer.Identifier && token.Value == typeKeyword
}

func (p *Parser) markStartPosition() {
	p.startPositions.Push(p.getStartPosition())
}

func (p *Parser) getStartPosition() uint {
	return uint(p.lexer.StartPosition())
}

func (p *Parser) getEndPosition() uint {
	return uint(p.lexer.BeforeWhitespacePosition())
}

func (p *Parser) error(location ast.Location, message string) {
	p.Diagnostics = append(p.Diagnostics,
		ast.NewDiagnostic(
			nil,
			location,
			message,
		),
	)
}

func (p *Parser) errorf(location ast.Location, format string, a ...any) {
	p.error(location, fmt.Sprintf(format, a...))
}

func (p *Parser) mark() ParserState {
	return ParserState{
		lexerState:        p.lexer.Mark(),
		startPositionsLen: p.startPositions.Size(),
		diagnosticLen:     len(p.Diagnostics),
	}
}

func (p *Parser) rewind(state ParserState) {
	p.lexer.Rewind(state.lexerState)
	p.startPositions.PopItems(p.startPositions.Size() - state.startPositionsLen)
	p.Diagnostics = p.Diagnostics[0:state.diagnosticLen]
}

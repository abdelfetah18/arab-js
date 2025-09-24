package parser

import (
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/lexer"
	"arab_js/internal/stack"
	"fmt"
)

type Parser struct {
	lexer          *lexer.Lexer
	startPositions stack.Stack[uint]

	Diagnostics []*ast.Diagnostic
}

func NewParser(lexer *lexer.Lexer) *Parser {
	return &Parser{
		lexer:          lexer,
		startPositions: stack.Stack[uint]{},
		Diagnostics:    []*ast.Diagnostic{},
	}
}

func ParseSourceFile(sourceText string) *ast.SourceFile {
	return NewParser(lexer.NewLexer(sourceText)).Parse()
}

func (p *Parser) Parse() *ast.SourceFile {
	p.markStartPosition()

	directives := p.parseDirectives()

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != lexer.EOF && p.lexer.Peek().Type != lexer.Invalid {
		statement := p.parseStatement()
		statements = append(statements, statement)
	}

	return ast.NewNode(ast.NewSourceFile(statements, directives), ast.Location{Pos: p.startPositions.Pop(), End: uint(p.lexer.Position())})
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

func (p *Parser) parseStatement() *ast.Node {
	token := p.lexer.Peek()

	switch token.Type {
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordIf:
			return p.parseIfStatement().AsNode()
		case lexer.KeywordLet:
			return p.parseVariableDeclaration().AsNode()
		case lexer.KeywordFunction:
			return p.parseFunctionDeclaration().AsNode()
		case lexer.KeywordImport:
			return p.parseImportDeclaration().AsNode()
		case lexer.KeywordExport:
			return p.parseExportDeclaration()
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
			p.markStartPosition()

			hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
			originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue
			p.expectedTypeKeyword(lexer.TypeKeywordDeclare)

			if p.optionalKeyword(lexer.KeywordLet) {
				identifier := p.parseIdentifier(true)
				if hasPrecedingOriginalNameDirective {
					identifier.OriginalName = &originalNameDirectiveValue
				}

				p.expected(lexer.Semicolon)

				return ast.NewNode(
					ast.NewVariableDeclaration(identifier, nil, true),
					ast.Location{
						Pos: p.startPositions.Pop(),
						End: p.getEndPosition(),
					},
				).AsNode()
			}

			p.errorf(
				ast.Location{
					Pos: p.startPositions.Pop(),
					End: p.getEndPosition(),
				},
				"Expected a variable declaration but got '%s'\n", p.lexer.Peek().Value)
			return nil
		}
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

func (p *Parser) parseImportDeclaration() *ast.ImportDeclaration {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordImport)

	importSpecifiers := []*ast.Node{}

	token := p.lexer.Peek()
	switch token.Type {
	case lexer.LeftCurlyBrace:
		importSpecifiers = p.parseImportSpecifiers(importSpecifiers)
	case lexer.Identifier:
		identifier := p.parseIdentifier(false)
		importSpecifiers = append(importSpecifiers,
			ast.NewNode(
				ast.NewImportDefaultSpecifier(identifier),
				identifier.Location,
			).AsNode(),
		)

		if p.optional(lexer.Comma) {
			token = p.lexer.Peek()
			if token.Type == lexer.LeftCurlyBrace {
				p.parseImportSpecifiers(importSpecifiers)
			}
		}
	case lexer.Star:
		p.expected(lexer.Star)
		p.expectedKeyword(lexer.KeywordAs)
		identifier := p.parseIdentifier(false)
		importSpecifiers = append(importSpecifiers,
			ast.NewNode(
				ast.NewImportNamespaceSpecifier(identifier),
				identifier.Location,
			).AsNode(),
		)
	default:
		p.errorf(
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
			"unexpected token got '%s'\n", token.Value)
		return nil
	}

	p.expectedKeyword(lexer.KeywordFrom)
	stringLiteral := p.parseStringLiteral()
	p.expected(lexer.Semicolon)

	return ast.NewNode(
		ast.NewImportDeclaration(importSpecifiers, stringLiteral),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseImportSpecifiers(importSpecifiers []*ast.Node) []*ast.Node {
	p.expected(lexer.LeftCurlyBrace)
	token := p.lexer.Peek()
	for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightCurlyBrace {
		identifier := p.parseIdentifier(false)
		importSpecifiers = append(
			importSpecifiers,
			ast.NewNode(
				ast.NewImportSpecifier(
					identifier,
					identifier.AsNode(),
				),
				identifier.Location,
			).AsNode(),
		)

		token = p.lexer.Peek()
		if token.Type != lexer.RightCurlyBrace {
			p.expected(lexer.Comma)
		}

		token = p.lexer.Peek()
	}

	p.expected(lexer.RightCurlyBrace)

	return importSpecifiers
}

func (p *Parser) parseExportNamedDeclaration() *ast.ExportNamedDeclaration {
	var declaration *ast.Node = nil
	specifiers := []*ast.ExportSpecifier{}
	var source *ast.StringLiteral = nil
	if p.optional(lexer.LeftCurlyBrace) {
		token := p.lexer.Peek()
		for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightCurlyBrace {
			identifier := p.parseIdentifier(false)
			specifiers = append(specifiers,
				ast.NewNode(
					ast.NewExportSpecifier(identifier, identifier.AsNode()),
					identifier.Location,
				),
			)

			if p.optional(lexer.Comma) {
				token = p.lexer.Peek()
			} else {
				p.expected(lexer.RightCurlyBrace)
			}
		}

		p.expected(lexer.RightCurlyBrace)

		if p.optionalKeyword(lexer.KeywordFrom) {
			source = p.parseStringLiteral()
		}
		p.expected(lexer.Semicolon)
	} else {
		declaration = p.parseDeclarationOnly()
	}

	return ast.NewNode(
		ast.NewExportNamedDeclaration(declaration, specifiers, source),
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

func (p *Parser) parseDeclarationOnly() *ast.Node {
	token := p.lexer.Peek()

	switch token.Type {
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordFunction:
			return p.parseFunctionDeclaration().AsNode()
		case lexer.KeywordLet, lexer.KeywordConst:
			return p.parseVariableDeclaration().AsNode()
		}
	}

	p.errorf(
		ast.Location{
			Pos: p.startPositions.Peek(),
			End: p.getEndPosition(),
		},
		"Unexpected token in export declaration: %s\n", token.Value)
	return nil
}

func (p *Parser) parseFunctionDeclarationOrExpression() *ast.Node {
	token := p.lexer.Peek()
	if token.Type == lexer.KeywordToken && token.Value == lexer.KeywordFunction {
		return p.parseFunctionDeclaration().AsNode()
	}

	return p.parseExpression()
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != lexer.EOF && p.lexer.Peek().Type != lexer.Invalid && p.lexer.Peek().Type != lexer.RightCurlyBrace {
		statements = append(statements, p.parseStatement())
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
		case lexer.KeywordFunction, lexer.KeywordLet, lexer.KeywordConst:
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
	case lexer.Identifier, lexer.SingleQuoteString, lexer.DoubleQuoteString, lexer.Decimal, lexer.LeftSquareBracket, lexer.LeftCurlyBrace:
		return true
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordNull, lexer.KeywordTrue, lexer.KeywordFalse:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (p *Parser) parsePrimaryExpression() *ast.Node {
	token := p.lexer.Peek()
	switch token.Type {
	case lexer.Identifier:
		return p.parseIdentifier(false).AsNode()
	case lexer.KeywordToken:
		switch token.Value {
		case lexer.KeywordNull:
			return p.parseNullLiteral().AsNode()
		case lexer.KeywordTrue, lexer.KeywordFalse:
			return p.parseBooleanLiteral().AsNode()
		}
	case lexer.Decimal:
		return p.parseDecimalLiteral().AsNode()
	case lexer.SingleQuoteString, lexer.DoubleQuoteString:
		return p.parseStringLiteral().AsNode()
	case lexer.LeftSquareBracket:
		return p.parseArrayExpression().AsNode()
	case lexer.LeftCurlyBrace:
		return p.parseObjectExpression().AsNode()
	}

	p.errorf(
		ast.Location{
			Pos: p.startPositions.Peek(),
			End: p.getEndPosition(),
		},
		"Expected PrimaryExpression but got %s\n", p.lexer.Peek().Value)
	return nil
}

func (p *Parser) parseTypeNode() *ast.Node {
	token := p.lexer.Peek()

	switch token.Value {
	case lexer.TypeKeywordString:
		return p.parseStringKeyword().AsNode()
	case lexer.TypeKeywordNumber:
		return p.parseNumberKeyword().AsNode()
	case lexer.TypeKeywordBoolean:
		return p.parseBooleanKeyword().AsNode()
	case lexer.TypeKeywordAny:
		return p.parseAnyKeyword().AsNode()

	default:
		return p.parseTypeReference().AsNode()
	}
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordLet)
	identifier := p.parseIdentifier(true)
	p.expected(lexer.Equal)
	assignmentExpression := p.parseAssignmentExpression()
	initializer := ast.NewNode(
		ast.NewInitializer(assignmentExpression),
		assignmentExpression.Location,
	)
	p.expected(lexer.Semicolon)

	return ast.NewNode(
		ast.NewVariableDeclaration(identifier, initializer, false),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordFunction)
	identifier := p.parseIdentifier(false)
	p.expected(lexer.LeftParenthesis)

	token := p.lexer.Peek()
	params := []*ast.Node{}

	isInLoop := func() bool {
		token := p.lexer.Peek()
		return token.Type != lexer.EOF &&
			token.Type != lexer.Invalid &&
			token.Type != lexer.RightParenthesis &&
			(token.Type == lexer.Identifier || token.Type == lexer.TripleDots)
	}

	for isInLoop() {
		pos := uint(p.lexer.Position())
		if p.optional(lexer.TripleDots) {
			identifier := p.parseIdentifier(false)
			var typeAnnotation *ast.TypeAnnotation = nil
			if p.optional(lexer.Colon) {
				typeAnnotation = p.parseTypeAnnotation()
			}

			params = append(params,
				ast.NewNode(
					ast.NewRestElement(identifier, typeAnnotation),
					ast.Location{
						Pos: pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)
		} else {
			params = append(params, p.parseIdentifier(true).AsNode())
		}

		token = p.lexer.Peek()
		if token.Type != lexer.RightParenthesis && token.Type != lexer.TripleDots {
			p.expected(lexer.Comma)
		}
	}

	p.expected(lexer.RightParenthesis)

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(lexer.Colon) {
		typeAnnotation = p.parseTypeAnnotation()
	}

	body := p.parseBlockStatement()

	return ast.NewNode(
		ast.NewFunctionDeclaration(identifier, params, body, typeAnnotation),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseLeftHandSideExpression() *ast.Node {
	expression := p.parseMemberExpression()

	for p.optional(lexer.LeftParenthesis) {
		argumentList := []*ast.Node{}

		token := p.lexer.Peek()
		for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightParenthesis {
			if !p.isExpression() {
				p.errorf(
					ast.Location{
						Pos: p.startPositions.Peek(),
						End: p.getEndPosition(),
					},
					"Expecting expression but got %s\n", token.Value)
			}
			argumentList = append(argumentList, p.parseExpression())
			token = p.lexer.Peek()
			if token.Type == lexer.Comma {
				token = p.lexer.Next()
			}
		}

		p.expected(lexer.RightParenthesis)
		expression = ast.NewNode(
			ast.NewCallExpression(expression, argumentList),
			ast.Location{
				Pos: expression.Location.Pos,
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	return expression
}

func (p *Parser) parseMemberExpression() *ast.Node {
	if p.isPrimaryExpression() {
		memberExpression := p.parsePrimaryExpression()

		isLeftSquareBracker := p.optional(lexer.LeftSquareBracket)
		for isLeftSquareBracker || p.optional(lexer.Dot) {
			var propertyNode *ast.Node = nil
			if isLeftSquareBracker {
				switch p.lexer.Peek().Type {
				case lexer.Identifier:
					propertyNode = p.parseIdentifier(false).AsNode()
				case lexer.DoubleQuoteString, lexer.SingleQuoteString:
					propertyNode = p.parseStringLiteral().AsNode()
				case lexer.Decimal:
					propertyNode = p.parseDecimalLiteral().AsNode()
				}
			} else {
				propertyNode = p.parseIdentifier(false).AsNode()
			}

			memberExpression = ast.NewNode(
				ast.NewMemberExpression(memberExpression, propertyNode, isLeftSquareBracker),
				ast.Location{
					Pos: memberExpression.Location.Pos,
					End: p.getEndPosition(),
				},
			).AsNode()
			if isLeftSquareBracker {
				p.expected(lexer.RightSquareBracket)
			}

			isLeftSquareBracker = p.optional(lexer.LeftSquareBracket)
		}

		return memberExpression
	}

	p.errorf(
		ast.Location{
			Pos: p.startPositions.Peek(),
			End: p.getEndPosition(),
		},
		"Expected MemberExpression but got %s\n", p.lexer.Peek().Value)
	return nil
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

	// Since its not update expression we remove position mark
	p.startPositions.Pop()

	return leftHandSideExpression
}

func (p *Parser) parseUnaryExpression() *ast.Node {
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
	return p.parseShortCircuitExpression()
}

func (p *Parser) parseAssignmentExpression() *ast.Node {
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
	return p.parseAssignmentExpression()
}

func (p *Parser) parseInterfaceDeclaration() *ast.InterfaceDeclaration {
	p.markStartPosition()

	p.expectedTypeKeyword(lexer.TypeKeywordInterface)

	identifier := p.parseIdentifier(false)
	body := p.parseInterfaceBody()

	p.optional(lexer.Semicolon)

	return ast.NewNode(
		ast.NewInterfaceDeclaration(identifier, body),
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

	token := p.lexer.Peek()
	for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightCurlyBrace {
		hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
		originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue

		var key *ast.Node
		switch token.Type {
		case lexer.Identifier:
			identifier := p.parseIdentifier(false)
			if hasPrecedingOriginalNameDirective {
				identifier.OriginalName = &originalNameDirectiveValue
			}

			key = identifier.AsNode()
		case lexer.DoubleQuoteString, lexer.SingleQuoteString:
			key = p.parseStringLiteral().AsNode()
		case lexer.Decimal:
			key = p.parseDecimalLiteral().AsNode()
		default:
			p.errorf(
				ast.Location{
					Pos: p.startPositions.Pop(),
					End: p.getEndPosition(),
				},
				"Expected a valid key but got %s\n", token.Value)
			return nil
		}

		if p.optional(lexer.Colon) {
			body = append(body,
				ast.NewNode(
					ast.NewPropertySignature(key, p.parseTypeAnnotation()),
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
					"Expected '{' but got %s\n", token.Value)
				return nil
			}

			p.optional(lexer.Comma)
			if key.Type == ast.NodeTypeIdentifier {
				body = append(body,
					ast.NewNode(
						ast.NewPropertySignature(key, p.parseTypeAnnotation()),
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
		ast.NewInterfaceBody(body),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	p.markStartPosition()

	switch p.lexer.Peek().Type {
	case lexer.LeftParenthesis:
		return ast.NewNode(
			ast.NewTypeAnnotation(
				p.parseFunctionType().AsNode(),
			),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		)
	case lexer.LeftCurlyBrace:
		return ast.NewNode(
			ast.NewTypeAnnotation(
				p.parseTypeLiteral().AsNode(),
			),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		)
	}

	typeNode := p.parseTypeNode()

	if p.optional(lexer.LeftSquareBracket) && p.optional(lexer.RightSquareBracket) {
		pos := p.startPositions.Pop()
		arrayType := ast.NewNode(
			ast.NewArrayType(typeNode),
			ast.Location{
				Pos: pos,
				End: p.getEndPosition(),
			},
		)

		return ast.NewNode(
			ast.NewTypeAnnotation(arrayType.AsNode()),
			ast.Location{
				Pos: pos,
				End: p.getEndPosition(),
			},
		)
	}

	return ast.NewNode(
		ast.NewTypeAnnotation(typeNode),
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

	p.expected(lexer.LeftParenthesis)

	params := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightParenthesis && (token.Type == lexer.Identifier || token.Type == lexer.TripleDots) {
		pos := uint(p.lexer.Position())
		if p.optional(lexer.TripleDots) {
			identifier := p.parseIdentifier(false)
			p.expected(lexer.Colon)
			typeAnnotation := p.parseTypeAnnotation()
			params = append(params,
				ast.NewNode(
					ast.NewRestElement(identifier, typeAnnotation),
					ast.Location{
						Pos: pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)
		} else {
			params = append(params, p.parseIdentifier(true).AsNode())
		}

		token = p.lexer.Peek()
		if token.Type != lexer.RightParenthesis {
			p.expected(lexer.Comma)
		}

		token = p.lexer.Peek()
	}

	p.expected(lexer.RightParenthesis)
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
	p.expected(lexer.Equal)
	typeAnnotation := p.parseTypeAnnotation()

	return ast.NewNode(
		ast.NewTypeAliasDeclaration(identifier, typeAnnotation),
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
		switch token.Type {
		case lexer.Identifier:
			key = p.parseIdentifier(false).AsNode()
		case lexer.DoubleQuoteString, lexer.SingleQuoteString:
			key = p.parseStringLiteral().AsNode()
		case lexer.Decimal:
			key = p.parseDecimalLiteral().AsNode()
		default:
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
					ast.NewPropertySignature(key, p.parseTypeAnnotation()),
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
						ast.NewPropertySignature(key, p.parseTypeAnnotation()),
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
	if token.Type == lexer.KeywordToken && token.Value == lexer.KeywordLet {
		init = p.parseVariableDeclaration().AsNode()
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

func (p *Parser) parseIdentifier(doParseTypeAnnotation bool) *ast.Identifier {
	p.markStartPosition()

	identifierName := p.lexer.Peek().Value
	if !p.expected(lexer.Identifier) {
		pos := p.startPositions.Pop()
		return ast.NewNode(
			ast.NewIdentifier(identifierName, nil),
			ast.Location{
				Pos: pos,
				End: pos,
			},
		)
	}

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

func (p *Parser) parseExportDeclaration() *ast.Node {
	p.markStartPosition()

	p.expectedKeyword(lexer.KeywordExport)

	if p.optionalKeyword(lexer.KeywordDefault) {
		return p.parseExportDefaultDeclaration().AsNode()
	}

	return p.parseExportNamedDeclaration().AsNode()
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

func (p *Parser) parseObjectExpression() *ast.ObjectExpression {
	p.markStartPosition()

	p.expected(lexer.LeftCurlyBrace)

	properties := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != lexer.EOF && token.Type != lexer.Invalid && token.Type != lexer.RightCurlyBrace {
		if p.optional(lexer.TripleDots) {
			pos := uint(p.lexer.Position())
			properties = append(properties,
				ast.NewNode(
					ast.NewSpreadElement(p.parseAssignmentExpression()),
					ast.Location{
						Pos: pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)

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
			token = p.lexer.Peek()
		} else {
			var key *ast.Node
			switch token.Type {
			case lexer.Identifier:
				key = p.parseIdentifier(false).AsNode()
			case lexer.DoubleQuoteString, lexer.SingleQuoteString:
				key = p.parseStringLiteral().AsNode()
			case lexer.Decimal:
				key = p.parseDecimalLiteral().AsNode()
			default:
				p.errorf(
					ast.Location{
						Pos: p.startPositions.Pop(),
						End: p.getEndPosition(),
					},
					"Expected a valid key but got %s\n", token.Value)
				return nil
			}

			if p.optional(lexer.Colon) {
				properties = append(properties,
					ast.NewNode(
						ast.NewObjectProperty(key, p.parseAssignmentExpression()),
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
					properties = append(properties,
						ast.NewNode(
							ast.NewObjectProperty(key, key),
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
		}
		token = p.lexer.Peek()
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

func (p *Parser) parseTypeReference() *ast.TypeReference {
	p.markStartPosition()

	identifier := p.parseIdentifier(false)

	return ast.NewNode(
		ast.NewTypeReference(identifier),
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

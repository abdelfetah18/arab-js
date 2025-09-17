package compiler

import (
	"arab_js/internal/compiler/ast"
	"arab_js/internal/stack"
)

type Parser struct {
	lexer          *Lexer
	startPositions stack.Stack[uint]
}

func NewParser(lexer *Lexer) *Parser {
	return &Parser{lexer: lexer, startPositions: stack.Stack[uint]{}}
}

func ParseSourceFile(sourceText string) *ast.SourceFile {
	return NewParser(NewLexer(sourceText)).Parse()
}

func (p *Parser) Parse() *ast.SourceFile {
	p.markStartPosition()

	directives := p.parseDirectives()

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != EOF && p.lexer.Peek().Type != Invalid {
		statements = append(statements, p.parseStatement())
	}

	token := p.lexer.Peek()
	if token.Type == Invalid {
		panic("Unexpected token: " + token.Value)
	}

	return ast.NewNode(ast.NewSourceFile(statements, directives), ast.Location{Pos: p.startPositions.Pop(), End: uint(p.lexer.position)})
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	p.markStartPosition()

	p.expectedKeyword(KeywordIf)
	p.expected(LeftParenthesis)
	testExpression := p.parseExpression()
	p.expected(RightParenthesis)
	consequentStatement := p.parseStatement()

	if p.optionalKeyword(KeywordElse) {
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
	case KeywordToken:
		switch token.Value {
		case KeywordIf:
			return p.parseIfStatement().AsNode()
		case KeywordLet:
			return p.parseVariableDeclaration().AsNode()
		case KeywordFunction:
			return p.parseFunctionDeclaration().AsNode()
		case KeywordImport:
			return p.parseImportDeclaration().AsNode()
		case KeywordExport:
			return p.parseExportDeclaration()
		case KeywordReturn:
			return p.parseReturnStatement().AsNode()
		case KeywordFor:
			return p.parseBreakableStatement()
		}
	case LeftCurlyBrace:
		return p.parseBlockStatement().AsNode()
	case Identifier:
		switch token.Value {
		case TypeKeywordInterface:
			return p.parseInterfaceDeclaration().AsNode()
		case TypeKeywordType:
			return p.parseTypeAliasDeclaration().AsNode()
		case TypeKeywordDeclare:
			p.markStartPosition()

			hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
			originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue

			p.expectedTypeKeyword(TypeKeywordDeclare)

			if p.optionalKeyword(KeywordLet) {
				identifier := p.parseIdentifier(true)
				if hasPrecedingOriginalNameDirective {
					identifier.OriginalName = &originalNameDirectiveValue
				}

				p.expected(Semicolon)

				return ast.NewNode(
					ast.NewVariableDeclaration(identifier, nil, true),
					ast.Location{
						Pos: p.startPositions.Pop(),
						End: p.getEndPosition(),
					},
				).AsNode()
			}

			panic("Expected a declaration but got '" + p.lexer.Peek().Value + "'")
		}
	}

	if p.isExpression() {
		p.markStartPosition()

		expression := p.parseExpression()
		p.expected(Semicolon)

		return ast.NewNode(
			ast.NewExpressionStatement(expression),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		).AsNode()
	}

	panic("Expected Statement got " + p.lexer.Peek().Value)
}

func (p *Parser) parseImportDeclaration() *ast.ImportDeclaration {
	p.markStartPosition()

	p.expectedKeyword(KeywordImport)

	importSpecifiers := []*ast.Node{}

	token := p.lexer.Peek()
	switch token.Type {
	case LeftCurlyBrace:
		p.parseImportSpecifiers(importSpecifiers)
	case Identifier:
		identifier := p.parseIdentifier(false)
		importSpecifiers = append(importSpecifiers,
			ast.NewNode(
				ast.NewImportDefaultSpecifier(identifier),
				identifier.Location,
			).AsNode(),
		)

		if p.optional(Comma) {
			token = p.lexer.Peek()
			if token.Type == LeftCurlyBrace {
				p.parseImportSpecifiers(importSpecifiers)
			}
		}
	case Star:
		p.expected(Star)
		p.expectedKeyword(KeywordAs)
		identifier := p.parseIdentifier(false)

		importSpecifiers = append(importSpecifiers,
			ast.NewNode(
				ast.NewImportNamespaceSpecifier(identifier),
				identifier.Location,
			).AsNode(),
		)
	default:
		panic("unexpected token got '" + token.Value + "'")
	}

	p.expectedKeyword(KeywordFrom)
	stringLiteral := p.parseStringLiteral()
	p.expected(Semicolon)

	return ast.NewNode(
		ast.NewImportDeclaration(importSpecifiers, stringLiteral),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseImportSpecifiers(importSpecifiers []*ast.Node) []*ast.Node {
	p.expected(LeftCurlyBrace)
	token := p.lexer.Peek()
	for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
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
		if token.Type != RightCurlyBrace {
			p.expected(Comma)
		}

		token = p.lexer.Peek()
	}

	p.expected(RightCurlyBrace)

	return importSpecifiers
}

func (p *Parser) parseExportNamedDeclaration() *ast.ExportNamedDeclaration {
	var declaration *ast.Node = nil
	specifiers := []*ast.ExportSpecifier{}
	var source *ast.StringLiteral = nil
	if p.optional(LeftCurlyBrace) {
		token := p.lexer.Peek()
		for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
			identifier := p.parseIdentifier(false)
			specifiers = append(specifiers,
				ast.NewNode(
					ast.NewExportSpecifier(identifier, identifier.AsNode()),
					identifier.Location,
				),
			)

			if p.optional(Comma) {
				token = p.lexer.Peek()
			} else {
				p.expected(RightCurlyBrace)
			}
		}

		p.expected(RightCurlyBrace)

		if p.optionalKeyword(KeywordFrom) {
			source = p.parseStringLiteral()
		}
		p.expected(Semicolon)
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
	p.expected(Semicolon)
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
	case KeywordToken:
		switch token.Value {
		case KeywordFunction:
			return p.parseFunctionDeclaration().AsNode()
		case KeywordLet, KeywordConst:
			return p.parseVariableDeclaration().AsNode()
		}
	}

	panic("Unexpected token in export declaration: " + token.Value)
}

func (p *Parser) parseFunctionDeclarationOrExpression() *ast.Node {
	token := p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == KeywordFunction {
		return p.parseFunctionDeclaration().AsNode()
	}

	return p.parseExpression()
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	p.markStartPosition()

	p.expected(LeftCurlyBrace)

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != EOF && p.lexer.Peek().Type != Invalid && p.lexer.Peek().Type != RightCurlyBrace {
		statements = append(statements, p.parseStatement())
	}

	p.expected(RightCurlyBrace)

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
	case LeftCurlyBrace:
		return false
	case KeywordToken:
		switch token.Value {
		case KeywordFunction, KeywordLet, KeywordConst:
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
	case Identifier, SingleQuoteString, DoubleQuoteString, Decimal, LeftSquareBracket, LeftCurlyBrace:
		return true
	case KeywordToken:
		switch token.Value {
		case KeywordNull, KeywordTrue, KeywordFalse:
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
	case Identifier:
		return p.parseIdentifier(false).AsNode()
	case KeywordToken:
		switch token.Value {
		case KeywordNull:
			return p.parseNullLiteral().AsNode()
		case KeywordTrue, KeywordFalse:
			return p.parseBooleanLiteral().AsNode()
		}
	case Decimal:
		return p.parseDecimalLiteral().AsNode()
	case SingleQuoteString, DoubleQuoteString:
		return p.parseStringLiteral().AsNode()
	case LeftSquareBracket:
		return p.parseArrayExpression().AsNode()
	case LeftCurlyBrace:
		return p.parseObjectExpression().AsNode()
	}

	panic("Expected PrimaryExpression but got " + p.lexer.Peek().Value)
}

func (p *Parser) parseTypeNode() *ast.Node {
	token := p.lexer.Peek()

	switch token.Value {
	case TypeKeywordString:
		return p.parseStringKeyword().AsNode()
	case TypeKeywordNumber:
		return p.parseNumberKeyword().AsNode()
	case TypeKeywordBoolean:
		return p.parseBooleanKeyword().AsNode()
	case TypeKeywordAny:
		return p.parseAnyKeyword().AsNode()

	default:
		return p.parseTypeReference().AsNode()
	}
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	p.markStartPosition()

	p.expectedKeyword(KeywordLet)
	identifier := p.parseIdentifier(true)
	p.expected(Equal)
	assignmentExpression := p.parseAssignmentExpression()
	initializer := ast.NewNode(
		ast.NewInitializer(assignmentExpression),
		assignmentExpression.Location,
	)
	p.expected(Semicolon)

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

	p.expectedKeyword(KeywordFunction)
	identifier := p.parseIdentifier(false)
	p.expected(LeftParenthesis)

	token := p.lexer.Peek()
	params := []*ast.Node{}

	isInLoop := func() bool {
		token := p.lexer.Peek()
		return token.Type != EOF &&
			token.Type != Invalid &&
			token.Type != RightParenthesis &&
			(token.Type == Identifier || token.Type == TripleDots)
	}

	for isInLoop() {
		pos := uint(p.lexer.position)
		if p.optional(TripleDots) {
			identifier := p.parseIdentifier(false)
			var typeAnnotation *ast.TypeAnnotation = nil
			if p.optional(Colon) {
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
		if token.Type != RightParenthesis && token.Type != TripleDots {
			p.expected(Comma)
		}
	}

	p.expected(RightParenthesis)

	var typeAnnotation *ast.TypeAnnotation = nil
	if p.optional(Colon) {
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

	for p.optional(LeftParenthesis) {
		argumentList := []*ast.Node{}

		token := p.lexer.Peek()
		for token.Type != EOF && token.Type != Invalid && token.Type != RightParenthesis {
			if !p.isExpression() {
				panic("Expecting *ast.Nodegot " + token.Value)
			}
			argumentList = append(argumentList, p.parseExpression())
			token = p.lexer.Peek()
			if token.Type == Comma {
				token = p.lexer.Next()
			}
		}

		p.expected(RightParenthesis)
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

		for p.optional(Dot) {
			identifier := p.parseIdentifier(false)
			memberExpression = ast.NewNode(
				ast.NewMemberExpression(memberExpression, identifier.AsNode()),
				ast.Location{
					Pos: memberExpression.Location.Pos,
					End: p.getEndPosition(),
				},
			).AsNode()
		}

		return memberExpression
	}

	panic("Expected MemberExpression but got " + p.lexer.Peek().Value)
}

func (p *Parser) parseUpdateExpression() *ast.Node {
	p.markStartPosition()

	isUpdateExpression := func() (string, bool) {
		return p.lexer.Peek().Value, (p.optional(DoublePlus) || p.optional(DoubleMinus))
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
	isUpdateExpression := func(tokenType TokenType) bool {
		return tokenType != Plus && tokenType != Minus
	}
	token := p.lexer.Peek()
	left := p.parseUnaryExpression()
	if !isUpdateExpression(token.Type) {
		return left
	}

	if p.optional(DoubleStar) {
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
		return p.lexer.Peek().Value, (p.optional(Slash) || p.optional(Percent) || p.optional(Star))
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
		return p.lexer.Peek().Value, (p.optional(Plus) || p.optional(Minus))
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
		return p.lexer.Peek().Value, (p.optional(DoubleLeftArrow) ||
			p.optional(DoubleRightArrow) ||
			p.optional(TripleRightArrow))
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
		return p.lexer.Peek().Value, (p.optional(LeftArrow) ||
			p.optional(RightArrow) ||
			p.optional(LeftArrowEqual) ||
			p.optional(RightArrowEqual))
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
		return p.lexer.Peek().Value, (p.optional(EqualEqual) ||
			p.optional(EqualEqualEqual) ||
			p.optional(NotEqual) ||
			p.optional(NotEqualEqual))
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
	for p.optional(BitwiseAnd) {
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
	for p.optional(BitwiseXor) {
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
	for p.optional(BitwiseOr) {
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
				Equal,
				StarEqual,
				SlashEqual,
				PercentEqual,
				PlusEqual,
				MinusEqual,
				BitwiseAndEqual,
				BitwiseXorEqual,
				BitwiseOrEqual,
				DoubleLeftArrowEqual,
				DoubleRightArrowEqual,
				DoubleStarEqual:
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

	p.expectedTypeKeyword(TypeKeywordInterface)

	identifier := p.parseIdentifier(false)
	body := p.parseInterfaceBody()

	p.optional(Semicolon)

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

	p.expected(LeftCurlyBrace)

	body := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
		hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
		originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue

		var key *ast.Node
		switch token.Type {
		case Identifier:
			identifier := p.parseIdentifier(false)
			if hasPrecedingOriginalNameDirective {
				identifier.OriginalName = &originalNameDirectiveValue
			}

			key = identifier.AsNode()
		case DoubleQuoteString, SingleQuoteString:
			key = p.parseStringLiteral().AsNode()
		case Decimal:
			key = p.parseDecimalLiteral().AsNode()
		default:
			panic("Expected a valid key but got " + token.Value)
		}

		if p.optional(Colon) {
			body = append(body,
				ast.NewNode(
					ast.NewPropertySignature(key, p.parseTypeAnnotation()),
					ast.Location{
						Pos: key.Location.Pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)

			p.optional(Comma)
			token = p.lexer.Peek()
		} else {
			if token.Type != Comma && token.Type != RightCurlyBrace {
				panic("A Expected '{' but got " + token.Value)
			}

			p.optional(Comma)
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
				panic("Expected Identifier but got " + token.Value)
			}
		}

		token = p.lexer.Peek()
	}

	p.expected(RightCurlyBrace)

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
	case LeftParenthesis:
		return ast.NewNode(
			ast.NewTypeAnnotation(
				p.parseFunctionType().AsNode(),
			),
			ast.Location{
				Pos: p.startPositions.Pop(),
				End: p.getEndPosition(),
			},
		)
	case LeftCurlyBrace:
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

	if p.optional(LeftSquareBracket) && p.optional(RightSquareBracket) {
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

	p.expectedKeyword(KeywordReturn)
	argument := p.parseExpression()
	p.expected(Semicolon)

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

	p.expected(LeftParenthesis)

	params := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != EOF && token.Type != Invalid && token.Type != RightParenthesis && (token.Type == Identifier || token.Type == TripleDots) {
		pos := uint(p.lexer.position)
		if p.optional(TripleDots) {
			identifier := p.parseIdentifier(false)
			p.expected(Colon)
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
		if token.Type != RightParenthesis {
			p.expected(Comma)
		}

		token = p.lexer.Peek()
	}

	p.expected(RightParenthesis)
	p.expected(EqualRightArrow)

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

	p.expectedTypeKeyword(TypeKeywordType)
	identifier := p.parseIdentifier(false)
	p.expected(Equal)
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

	p.expected(LeftCurlyBrace)

	members := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
		var key *ast.Node
		switch token.Type {
		case Identifier:
			key = p.parseIdentifier(false).AsNode()
		case DoubleQuoteString, SingleQuoteString:
			key = p.parseStringLiteral().AsNode()
		case Decimal:
			key = p.parseDecimalLiteral().AsNode()
		default:
			panic("Expected a valid key but got " + token.Value)
		}

		if p.optional(Colon) {
			members = append(members,
				ast.NewNode(
					ast.NewPropertySignature(key, p.parseTypeAnnotation()),
					ast.Location{
						Pos: key.Location.Pos,
						End: p.getEndPosition(),
					},
				).AsNode(),
			)

			p.optional(Comma)
			token = p.lexer.Peek()
		} else {
			if token.Type != Comma && token.Type != RightCurlyBrace {
				panic("A Expected '{' but got " + token.Value)
			}
			p.optional(Comma)
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
				panic("Expected Identifier but got " + token.Value)
			}
		}

		token = p.lexer.Peek()
	}

	p.expected(RightCurlyBrace)

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
	case KeywordFor:
		return p.parseIterationStatement()
	}

	return nil
}

func (p *Parser) parseIterationStatement() *ast.Node {
	token := p.lexer.Peek()
	switch token.Value {
	case KeywordFor:
		return p.parseForStatement().AsNode()
	}

	return nil
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	p.markStartPosition()

	p.expectedKeyword(KeywordFor)
	p.expected(LeftParenthesis)

	var init *ast.Node = nil
	token := p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == KeywordLet {
		init = p.parseVariableDeclaration().AsNode()
	} else {
		init = p.parseExpression()
		p.expected(Semicolon)
	}

	test := p.parseExpression()
	p.expected(Semicolon)
	update := p.parseExpression()
	p.expected(RightParenthesis)

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
	p.expected(Identifier)

	var typeAnnotation *ast.TypeAnnotation = nil
	if doParseTypeAnnotation && p.optional(Colon) {
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

	if !p.optional(DoubleQuoteString) && !p.optional(SingleQuoteString) {
		panic("Expected string but got '" + p.lexer.Peek().Value + "'")
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

	p.expectedKeyword(KeywordNull)

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
	p.expected(Decimal)

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
	if !p.optionalKeyword(KeywordTrue) && !p.optionalKeyword(KeywordFalse) {
		panic("Expected boolean but got '" + p.lexer.Peek().Value + "'")
	}

	return ast.NewNode(
		ast.NewBooleanLiteral(value == KeywordTrue),
		ast.Location{
			Pos: p.startPositions.Pop(),
			End: p.getEndPosition(),
		},
	)
}

func (p *Parser) parseExportDeclaration() *ast.Node {
	p.markStartPosition()

	p.expectedKeyword(KeywordExport)

	if p.optionalKeyword(KeywordDefault) {
		return p.parseExportDefaultDeclaration().AsNode()
	}

	return p.parseExportNamedDeclaration().AsNode()
}

func (p *Parser) parseDirective() *ast.Directive {
	p.markStartPosition()

	stringLiteralValue := p.lexer.Peek().Value

	if !p.optional(DoubleQuoteString) && !p.optional(SingleQuoteString) {
		panic("Expected string but got '" + p.lexer.Peek().Value + "'")
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
	for token.Type == SingleQuoteString || token.Type == DoubleQuoteString {
		directives = append(directives, p.parseDirective())
		p.expected(Semicolon)
		token = p.lexer.Peek()
	}

	return directives
}

func (p *Parser) parseArrayExpression() *ast.ArrayExpression {
	p.markStartPosition()

	p.expected(LeftSquareBracket)

	elements := []*ast.Node{}
	token := p.lexer.Peek()

	for token.Type != EOF && token.Type != Invalid && token.Type != RightSquareBracket {
		for p.optional(Comma) {
		}

		pos := uint(p.lexer.position)
		if p.optional(TripleDots) {
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

	p.expected(RightSquareBracket)

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

	p.expected(LeftCurlyBrace)

	properties := []*ast.Node{}

	token := p.lexer.Peek()
	for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
		if p.optional(TripleDots) {
			pos := uint(p.lexer.position)
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
			if token.Type != Comma && token.Type != RightCurlyBrace {
				panic("Expected '{' but got " + token.Value)
			}

			p.optional(Comma)
			token = p.lexer.Peek()
		} else {
			var key *ast.Node
			switch token.Type {
			case Identifier:
				key = p.parseIdentifier(false).AsNode()
			case DoubleQuoteString, SingleQuoteString:
				key = p.parseStringLiteral().AsNode()
			case Decimal:
				key = p.parseDecimalLiteral().AsNode()
			default:
				panic("Expected a valid key but got " + token.Value)
			}

			if p.optional(Colon) {
				properties = append(properties,
					ast.NewNode(
						ast.NewObjectProperty(key, p.parseAssignmentExpression()),
						ast.Location{
							Pos: key.Location.Pos,
							End: p.getEndPosition(),
						},
					).AsNode(),
				)

				p.optional(Comma)
				token = p.lexer.Peek()
			} else {
				if token.Type != Comma && token.Type != RightCurlyBrace {
					panic("A Expected '{' but got " + token.Value)
				}

				p.optional(Comma)
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
					panic("Expected Identifier but got " + token.Value)
				}
			}
		}
		token = p.lexer.Peek()
	}
	p.expected(RightCurlyBrace)

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

	p.expectedTypeKeyword(TypeKeywordString)

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

	p.expectedTypeKeyword(TypeKeywordNumber)

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

	p.expectedTypeKeyword(TypeKeywordBoolean)

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

	p.expectedTypeKeyword(TypeKeywordAny)

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

func (p *Parser) expected(tokenType TokenType) {
	token := p.lexer.Peek()
	if token.Type != tokenType {
		panic("expected '" + tokenType.String() + "' but got '" + token.Type.String() + "'")
	}
	p.lexer.Next()
}

func (p *Parser) expectedKeyword(keyword Keyword) {
	token := p.lexer.Peek()
	if token.Type != KeywordToken && token.Value != keyword {
		panic("expected '" + keyword + "' keyword but got '" + token.Value + "'")
	}
	p.lexer.Next()
}

func (p *Parser) expectedTypeKeyword(typeKeyword TypeKeyword) {
	token := p.lexer.Peek()
	if token.Type != Identifier && token.Value != typeKeyword {
		panic("expected '" + typeKeyword + "' type keyword but got '" + token.Value + "'")
	}
	p.lexer.Next()
}

func (p *Parser) optional(tokenType TokenType) bool {
	token := p.lexer.Peek()
	if token.Type == tokenType {
		p.lexer.Next()
		return true
	}
	return false
}

func (p *Parser) optionalKeyword(keyword Keyword) bool {
	token := p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == keyword {
		p.lexer.Next()
		return true
	}
	return false
}

func (p *Parser) markStartPosition() {
	p.startPositions.Push(p.getStartPosition())
}

func (p *Parser) getStartPosition() uint {
	return uint(p.lexer.startPosition)
}

func (p *Parser) getEndPosition() uint {
	return uint(p.lexer.beforeWhitespacePosition)
}

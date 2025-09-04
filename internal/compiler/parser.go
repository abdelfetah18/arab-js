package compiler

import (
	"arab_js/internal/compiler/ast"
)

type Parser struct {
	lexer          *Lexer
	isNativeScript bool
}

func NewParser(lexer *Lexer, isNativeScript bool) *Parser {
	return &Parser{lexer: lexer, isNativeScript: isNativeScript}
}

func ParseSourceFile(sourceText string) *ast.SourceFile {
	return NewParser(NewLexer(sourceText), false).Parse()
}

func (p *Parser) Parse() *ast.SourceFile {
	token := p.lexer.Peek()
	directives := []*ast.Directive{}
	for token.Type == SingleQuoteString || token.Type == DoubleQuoteString {
		value := token.Value[1 : len(token.Value)-1]
		if value == "أصلي" {
			p.isNativeScript = true
		}
		directives = append(directives, ast.NewDirective(ast.NewDirectiveLiteral(value)))
		token = p.lexer.Next()
		if token.Type != Semicolon {
			panic("Expected '؛', got " + token.Value)
		}
		token = p.lexer.Next()
	}

	statements := []*ast.Node{}
	for p.lexer.Peek().Type != EOF && p.lexer.Peek().Type != Invalid {
		statements = append(statements, p.parseStatement())
	}
	token = p.lexer.Peek()
	if token.Type == Invalid {
		panic("Unexpected token: " + token.Value)
	}

	return ast.NewSourceFile(statements, directives)
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	token := p.lexer.Peek()
	if token.Type != LeftParenthesis {
		panic("Expected ')' but got " + token.Value)
	}
	p.lexer.Next()
	testExpression := p.parseExpression()
	token = p.lexer.Peek()
	if token.Type != RightParenthesis {
		panic("Expected '(' but got " + token.Value)
	}
	p.lexer.Next()
	consequentStatement := p.parseStatement()
	token = p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == KeywordElse {
		p.lexer.Next()
		alternateStatement := p.parseStatement()
		return ast.NewIfStatement(testExpression, consequentStatement, alternateStatement)
	}
	return ast.NewIfStatement(testExpression, consequentStatement, nil)
}

func (p *Parser) parseStatement() *ast.Node {
	token := p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == KeywordIf {
		p.lexer.Next()
		return p.parseIfStatement().ToNode()
	}
	if token.Type == KeywordToken && token.Value == KeywordLet {
		p.lexer.Next()
		return p.parseVariableDeclaration().ToNode()
	}
	if token.Type == KeywordToken && token.Value == KeywordFunction {
		p.lexer.Next()
		return p.parseFunctionDeclaration().ToNode()
	}
	if token.Type == KeywordToken && token.Value == KeywordImport {
		p.lexer.Next()
		return p.parseImportDeclaration().ToNode()
	}
	if token.Type == KeywordToken && token.Value == KeywordExport {
		token = p.lexer.Next()
		if token.Type == KeywordToken && token.Value == KeywordDefault {
			p.lexer.Next()
			return p.parseExportDefaultDeclaration().ToNode()
		}
		return p.parseExportNamedDeclaration().ToNode()
	}
	if token.Type == KeywordToken && token.Value == KeywordReturn {
		p.lexer.Next()
		return p.parseReturnStatement().ToNode()
	}
	if token.Type == LeftCurlyBrace {
		p.lexer.Next()
		return p.parseBlockStatement().ToNode()
	}

	if token.Type == Identifier && token.Value == TypeKeywordInterface {
		p.lexer.Next()
		return p.parseTInterfaceDeclaration().ToNode()
	}

	if token.Type == Identifier && token.Value == TypeKeywordType {
		p.lexer.Next()
		return p.parseTTypeAliasDeclaration().ToNode()
	}

	if token.Type == Identifier && token.Value == TypeKeywordDeclare {
		hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
		originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue

		token := p.lexer.Next()
		if token.Type == KeywordToken && token.Value == KeywordLet {
			p.lexer.Next()
			identifier := p.parseTypedIdentifier()
			if hasPrecedingOriginalNameDirective {
				identifier.OriginalName = &originalNameDirectiveValue
			}

			token := p.lexer.Peek()
			if token.Type != Semicolon {
				panic("Expected '؛', got " + token.Value)
			}
			p.lexer.Next()

			return ast.NewVariableDeclaration(identifier, nil, true).ToNode()
		}
	}

	if p.isExpression() {
		expression := p.parseExpression()
		token := p.lexer.Peek()
		if token.Type != Semicolon {
			panic("Expected '؛', got " + token.Value)
		}
		p.lexer.Next()
		return ast.NewExpressionStatement(expression).ToNode()
	}
	panic("Expected Statement got " + p.lexer.Peek().Value)
}

func (p *Parser) parseImportDeclaration() *ast.ImportDeclaration {
	token := p.lexer.Peek()
	importSpecifiers := []ast.ImportSpecifierInterface{}
	if token.Type == LeftCurlyBrace {
		p.lexer.Next()
		token = p.lexer.Peek()
		for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
			if token.Type != Identifier {
				panic("Expected Identifier got " + token.Value)
			}
			importSpecifiers = append(
				importSpecifiers,
				ast.NewImportSpecifier(ast.NewIdentifier(token.Value, nil), ast.NewIdentifier(token.Value, nil).ToNode()),
			)
			p.lexer.Next()
			token = p.lexer.Peek()
			if token.Type == Comma {
				p.lexer.Next()
				token = p.lexer.Peek()
			} else if token.Type != RightCurlyBrace {
				panic("Expected '}' got " + token.Value)
			}
		}
		p.lexer.Next()
	} else if token.Type == Identifier {
		importSpecifiers = append(importSpecifiers, ast.NewImportDefaultSpecifier(ast.NewIdentifier(token.Value, nil)))
		p.lexer.Next()
		token = p.lexer.Peek()
		if token.Type == Comma {
			p.lexer.Next()
			token = p.lexer.Peek()
			if token.Type == LeftCurlyBrace {
				p.lexer.Next()
				token = p.lexer.Peek()
				for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
					if token.Type != Identifier {
						panic("Expected Identifier got " + token.Value)
					}
					importSpecifiers = append(
						importSpecifiers,
						ast.NewImportSpecifier(ast.NewIdentifier(token.Value, nil), ast.NewIdentifier(token.Value, nil).ToNode()),
					)
					p.lexer.Next()
					token = p.lexer.Peek()
					if token.Type == Comma {
						p.lexer.Next()
						token = p.lexer.Peek()
					} else if token.Type != RightCurlyBrace {
						panic("Expected '}' got " + token.Value)
					}
				}
				p.lexer.Next()
			}
		}
	} else if token.Type == Star {
		p.lexer.Next()
		token = p.lexer.Peek()
		if !(token.Type == KeywordToken && token.Value == KeywordAs) {
			panic("Expected 'as' after '*' got " + token.Value)
		}
		p.lexer.Next()
		token = p.lexer.Peek()
		if token.Type != Identifier {
			panic("Expected Identifier after 'as' got " + token.Value)
		}
		importSpecifiers = append(importSpecifiers, ast.NewImportNamespaceSpecifier(ast.NewIdentifier(token.Value, nil)))
		p.lexer.Next()
	} else {
		panic("Unexpected token in import: " + token.Value)
	}
	token = p.lexer.Peek()
	if !(token.Type == KeywordToken && token.Value == KeywordFrom) {
		panic("Expected 'من' got " + token.Value)
	}
	p.lexer.Next()
	moduleToken := p.lexer.Peek()
	if moduleToken.Type != DoubleQuoteString && moduleToken.Type != SingleQuoteString {
		panic("Expected StringLiteral got " + moduleToken.Value)
	}
	sourceLiteral := ast.NewStringLiteral(moduleToken.Value)
	p.lexer.Next()
	token = p.lexer.Peek()
	if token.Type != Semicolon {
		panic("Expected '؛' got " + token.Value)
	}
	p.lexer.Next()
	return ast.NewImportDeclaration(importSpecifiers, sourceLiteral)
}

func (p *Parser) parseExportNamedDeclaration() *ast.ExportNamedDeclaration {
	token := p.lexer.Peek()
	var declaration *ast.Node = nil
	specifiers := []*ast.ExportSpecifier{}
	var source *ast.StringLiteral = nil
	if token.Type == LeftCurlyBrace {
		p.lexer.Next()
		token = p.lexer.Peek()
		for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
			if token.Type != Identifier {
				panic("Expected Identifier in export got " + token.Value)
			}
			id := ast.NewIdentifier(token.Value, nil)
			specifiers = append(specifiers, ast.NewExportSpecifier(id, id.ToNode()))
			p.lexer.Next()
			token = p.lexer.Peek()
			if token.Type == Comma {
				p.lexer.Next()
				token = p.lexer.Peek()
			} else if token.Type != RightCurlyBrace {
				panic("Expected '}' or ',' got " + token.Value)
			}
		}
		if token.Type != RightCurlyBrace {
			panic("Expected '}' got " + token.Value)
		}
		p.lexer.Next()
		token = p.lexer.Peek()
		if token.Type == KeywordToken && token.Value == KeywordFrom {
			p.lexer.Next()
			stringToken := p.lexer.Peek()
			if stringToken.Type != DoubleQuoteString && stringToken.Type != SingleQuoteString {
				panic("Expected StringLiteral got " + stringToken.Value)
			}
			source = ast.NewStringLiteral(stringToken.Value)
			p.lexer.Next()
		}
		token = p.lexer.Peek()
		if token.Type != Semicolon {
			panic("Expected '؛' at end of export got " + token.Value)
		}
		p.lexer.Next()
	} else {
		declaration = p.parseDeclarationOnly()
	}
	return ast.NewExportNamedDeclaration(declaration, specifiers, source)
}

func (p *Parser) parseExportDefaultDeclaration() *ast.ExportDefaultDeclaration {
	declOrExpr := p.parseFunctionDeclarationOrExpression()
	token := p.lexer.Peek()
	if token.Type != Semicolon {
		panic("Expected '؛' at end of export default got " + token.Value)
	}
	p.lexer.Next()
	return ast.NewExportDefaultDeclaration(declOrExpr)
}

func (p *Parser) parseDeclarationOnly() *ast.Node {
	token := p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == KeywordFunction {
		p.lexer.Next()
		return p.parseFunctionDeclaration().ToNode()
	}
	if token.Type == KeywordToken && (token.Value == KeywordLet || token.Value == KeywordConst) {
		p.lexer.Next()
		return p.parseVariableDeclaration().ToNode()
	}
	panic("Unexpected token in export declaration: " + token.Value)
}

func (p *Parser) parseFunctionDeclarationOrExpression() *ast.Node {
	token := p.lexer.Peek()
	if token.Type == KeywordToken && token.Value == KeywordFunction {
		p.lexer.Next()
		return p.parseFunctionDeclaration().ToNode()
	}
	return p.parseExpression()
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	statements := []*ast.Node{}
	for p.lexer.Peek().Type != EOF && p.lexer.Peek().Type != Invalid && p.lexer.Peek().Type != RightCurlyBrace {
		statements = append(statements, p.parseStatement())
	}
	token := p.lexer.Peek()
	if token.Type == Invalid {
		panic("Unexpected token: " + token.Value)
	}
	if token.Type == EOF {
		panic("Expecting '{'")
	}
	if token.Type != RightCurlyBrace {
		panic("Expected '{' but got " + token.Value)
	}
	p.lexer.Next()
	return ast.NewBlockStatement(statements)
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
	if token.Type == Identifier {
		p.lexer.Next()
		return ast.NewIdentifier(token.Value, nil).ToNode()
	}
	if token.Type == KeywordToken && token.Value == KeywordNull {
		p.lexer.Next()
		return ast.NewNullLiteral().ToNode()
	}
	if token.Type == KeywordToken && (token.Value == KeywordTrue || token.Value == KeywordFalse) {
		p.lexer.Next()
		return ast.NewBooleanLiteral(token.Value == KeywordTrue).ToNode()
	}
	if token.Type == Decimal {
		p.lexer.Next()
		return ast.NewDecimalLiteral(token.Value).ToNode()
	}
	if token.Type == SingleQuoteString {
		p.lexer.Next()
		return ast.NewStringLiteral(token.Value[1 : len(token.Value)-1]).ToNode()
	}
	if token.Type == DoubleQuoteString {
		p.lexer.Next()
		return ast.NewStringLiteral(token.Value[1 : len(token.Value)-1]).ToNode()
	}
	if token.Type == LeftSquareBracket {
		token = p.lexer.Next()
		elements := []*ast.Node{}
		for token.Type != EOF && token.Type != Invalid && token.Type != RightSquareBracket {
			for token.Type == Comma {
				token = p.lexer.Next()
			}
			if token.Type == TripleDots {
				token = p.lexer.Next()
				elements = append(elements, ast.NewSpreadElement(p.parseAssignmentExpression()).ToNode())
			} else {
				elements = append(elements, p.parseAssignmentExpression())
			}
			token = p.lexer.Peek()
		}
		if token.Type != RightSquareBracket {
			panic("Expected '[' but got " + token.Value)
		}
		p.lexer.Next()
		return ast.NewArrayExpression(elements).ToNode()
	}
	if token.Type == LeftCurlyBrace {
		token = p.lexer.Next()
		properties := []*ast.Node{}
		for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
			if token.Type == TripleDots {
				token = p.lexer.Next()
				properties = append(properties, ast.NewSpreadElement(p.parseAssignmentExpression()).ToNode())
				token = p.lexer.Peek()
				if token.Type != Comma && token.Type != RightCurlyBrace {
					panic("Expected '{' but got " + token.Value)
				}
				if token.Type == Comma {
					token = p.lexer.Next()
				}
			} else {
				var key *ast.Node
				switch token.Type {
				case Identifier:
					key = ast.NewIdentifier(token.Value, nil).ToNode()
					token = p.lexer.Next()
				case DoubleQuoteString, SingleQuoteString:
					key = ast.NewStringLiteral(token.Value).ToNode()
					token = p.lexer.Next()
				case Decimal:
					key = ast.NewDecimalLiteral(token.Value).ToNode()
					token = p.lexer.Next()
				default:
					panic("Expected a valid key but got " + token.Value)
				}

				if token.Type == Colon {
					token = p.lexer.Next()
					properties = append(properties, ast.NewObjectProperty(key, p.parseAssignmentExpression()).ToNode())
					token = p.lexer.Peek()
					if token.Type == Comma {
						token = p.lexer.Next()
					}
				} else {
					if token.Type != Comma && token.Type != RightCurlyBrace {
						panic("A Expected '{' but got " + token.Value)
					}
					if token.Type == Comma {
						token = p.lexer.Next()
					}
					if key.Type == ast.NodeTypeIdentifier {
						properties = append(properties, ast.NewObjectProperty(key, key).ToNode())
					} else {
						panic("Expected Identifier but got " + token.Value)
					}
				}
			}
			token = p.lexer.Peek()
		}
		if token.Type != RightCurlyBrace {
			panic("Expected '{' but got " + token.Value)
		}
		p.lexer.Next()
		return ast.NewObjectExpression(properties).ToNode()
	}
	panic("Expected PrimaryExpression but got " + p.lexer.Peek().Value)
}

func (p *Parser) getTypeNodeFromIdentifier(token Token) *ast.Node {
	if token.Value == TypeKeywordString {
		return ast.NewTStringKeyword().ToNode()
	}

	if token.Value == TypeKeywordNumber {
		return ast.NewTNumberKeyword().ToNode()
	}

	if token.Value == TypeKeywordBoolean {
		return ast.NewTBooleanKeyword().ToNode()
	}

	if token.Value == TypeKeywordAny {
		return ast.NewTAnyKeyword().ToNode()
	}

	return ast.NewTTypeReference(ast.NewIdentifier(token.Value, nil)).ToNode()
}

func (p *Parser) parseVariableDeclaration() *ast.VariableDeclaration {
	identifier := p.parseTypedIdentifier()

	token := p.lexer.Peek()
	if token.Type != Equal {
		panic("Expected '=', got " + token.Value)
	}
	token = p.lexer.Next()

	init := ast.NewInitializer(p.parseAssignmentExpression())
	semicolon := p.lexer.Peek()
	if semicolon.Type != Semicolon {
		panic("Expected '؛', got " + semicolon.Value)
	}
	p.lexer.Next()
	return ast.NewVariableDeclaration(identifier, init, false)
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	token := p.lexer.Peek()
	if token.Type != Identifier {
		panic("Expected 'Identifier', got " + token.Value)
	}
	identifier := ast.NewIdentifier(token.Value, nil)
	token = p.lexer.Next()
	if token.Type != LeftParenthesis {
		panic("Expected ')', got " + token.Value)
	}
	token = p.lexer.Next()
	params := []*ast.Identifier{}
	for token.Type != EOF && token.Type != Invalid && token.Type != RightParenthesis && token.Type == Identifier {
		params = append(params, p.parseTypedIdentifier())
		token = p.lexer.Peek()
		if token.Type != Comma && token.Type != RightParenthesis {
			panic("Expected ',', got " + token.Value)
		}
		if token.Type == Comma {
			token = p.lexer.Next()
		}
	}
	if token.Type != RightParenthesis {
		panic("Expected '(', got " + token.Value)
	}
	token = p.lexer.Next()

	var tTypeAnnotation *ast.TTypeAnnotation = nil
	if token.Type == Colon {
		p.lexer.Next()
		tTypeAnnotation = p.parseTTypeAnnotation()
	}

	token = p.lexer.Peek()
	if token.Type != LeftCurlyBrace {
		panic("Expecting '}' got " + token.Value)
	}
	p.lexer.Next()
	body := p.parseBlockStatement()
	return ast.NewFunctionDeclaration(identifier, params, body, tTypeAnnotation)
}

func (p *Parser) parseLeftHandSideExpression() *ast.Node {
	expression := p.parseMemberExpression()
	token := p.lexer.Peek()
	for token.Type == LeftParenthesis {
		token = p.lexer.Next()
		argumentList := []*ast.Node{}
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
		if token.Type != RightParenthesis {
			panic("Expecting '(' got " + token.Value)
		}
		token = p.lexer.Next()
		expression = ast.NewCallExpression(expression, argumentList).ToNode()
	}
	return expression
}

func (p *Parser) parseMemberExpression() *ast.Node {
	if p.isPrimaryExpression() {
		memberExpression := p.parsePrimaryExpression()
		token := p.lexer.Peek()
		for token.Type == Dot {
			token = p.lexer.Next()
			if token.Type != Identifier {
				panic("Expected Identifier but got " + token.Value)
			}
			memberExpression = ast.NewMemberExpression(memberExpression, ast.NewIdentifier(token.Value, nil).ToNode()).ToNode()
			token = p.lexer.Next()
		}
		return memberExpression
	}
	panic("Expected MemberExpression but got " + p.lexer.Peek().Value)
}

func (p *Parser) parseUpdateExpression() *ast.Node {
	return p.parseLeftHandSideExpression()
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
	if p.lexer.Peek().Type == DoubleStar {
		p.lexer.Next()
		right := p.parseExponentiationExpression()
		left = ast.NewBinaryExpression("**", left, right).ToNode()
	}
	return left
}

func (p *Parser) parseMultiplicativeExpression() *ast.Node {
	isMultiplicativeToken := func(tokenType TokenType) bool {
		return tokenType == Slash || tokenType == Percent || tokenType == Star
	}
	left := p.parseExponentiationExpression()
	for isMultiplicativeToken(p.lexer.Peek().Type) {
		operator := p.lexer.Peek().Value
		p.lexer.Next()
		right := p.parseExponentiationExpression()
		left = ast.NewBinaryExpression(operator, left, right).ToNode()
	}
	return left
}

func (p *Parser) parseAdditiveExpression() *ast.Node {
	isAdditiveToken := func(tokenType TokenType) bool {
		return tokenType == Plus || tokenType == Minus
	}
	left := p.parseMultiplicativeExpression()
	for isAdditiveToken(p.lexer.Peek().Type) {
		operator := p.lexer.Peek().Value
		p.lexer.Next()
		right := p.parseMultiplicativeExpression()
		left = ast.NewBinaryExpression(operator, left, right).ToNode()
	}
	return left
}

func (p *Parser) parseShiftExpression() *ast.Node {
	isShiftToken := func(tokenType TokenType) bool {
		return tokenType == DoubleLeftArrow || tokenType == DoubleRightArrow || tokenType == TripleRightArrow
	}
	left := p.parseAdditiveExpression()
	for isShiftToken(p.lexer.Peek().Type) {
		operator := p.lexer.Peek().Value
		p.lexer.Next()
		right := p.parseAdditiveExpression()
		left = ast.NewBinaryExpression(operator, left, right).ToNode()
	}
	return left
}

func (p *Parser) parseRelationalExpression() *ast.Node {
	isRelationalToken := func(tokenType TokenType) bool {
		return tokenType == LeftArrow || tokenType == RightArrow || tokenType == LeftArrowEqual || tokenType == RightArrowEqual
	}
	left := p.parseShiftExpression()
	for isRelationalToken(p.lexer.Peek().Type) {
		operator := p.lexer.Peek().Value
		p.lexer.Next()
		right := p.parseShiftExpression()
		left = ast.NewBinaryExpression(operator, left, right).ToNode()
	}
	return left
}

func (p *Parser) parseEqualityExpression() *ast.Node {
	isEqualityToken := func(tokenType TokenType) bool {
		return tokenType == EqualEqual || tokenType == EqualEqualEqual || tokenType == NotEqual || tokenType == NotEqualEqual
	}
	left := p.parseRelationalExpression()
	for isEqualityToken(p.lexer.Peek().Type) {
		operator := p.lexer.Peek().Value
		p.lexer.Next()
		right := p.parseRelationalExpression()
		left = ast.NewBinaryExpression(operator, left, right).ToNode()
	}
	return left
}

func (p *Parser) parseBitwiseANDExpression() *ast.Node {
	left := p.parseEqualityExpression()
	for p.lexer.Peek().Type == BitwiseAnd {
		p.lexer.Next()
		right := p.parseEqualityExpression()
		left = ast.NewBinaryExpression("&", left, right).ToNode()
	}
	return left
}

func (p *Parser) parseBitwiseXORExpression() *ast.Node {
	left := p.parseBitwiseANDExpression()
	for p.lexer.Peek().Type == BitwiseXor {
		p.lexer.Next()
		right := p.parseBitwiseANDExpression()
		left = ast.NewBinaryExpression("^", left, right).ToNode()
	}
	return left
}

func (p *Parser) parseBitwiseORExpression() *ast.Node {
	left := p.parseBitwiseXORExpression()
	for p.lexer.Peek().Type == BitwiseOr {
		p.lexer.Next()
		right := p.parseBitwiseXORExpression()
		left = ast.NewBinaryExpression("|", left, right).ToNode()
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
	return p.parseConditionalExpression()
}

func (p *Parser) parseExpression() *ast.Node {
	return p.parseAssignmentExpression()
}

func (p *Parser) parseTInterfaceDeclaration() *ast.TInterfaceDeclaration {
	token := p.lexer.Peek()
	if token.Type != Identifier {
		panic("Expected Identifier got " + token.Value)
	}

	id := ast.NewIdentifier(token.Value, nil)
	token = p.lexer.Next()

	body := p.parseTInterfaceBody()

	return ast.NewTInterfaceDeclaration(id, body)
}

func (p *Parser) parseTInterfaceBody() *ast.TInterfaceBody {
	token := p.lexer.Peek()
	if token.Type != LeftCurlyBrace {
		panic("Expected '}' got " + token.Value)
	}
	token = p.lexer.Next()

	body := []*ast.Node{}
	for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
		hasPrecedingOriginalNameDirective := p.lexer.HasPrecedingOriginalNameDirective
		originalNameDirectiveValue := p.lexer.OriginalNameDirectiveValue

		var key *ast.Node
		switch token.Type {
		case Identifier:
			identifier := ast.NewIdentifier(token.Value, nil)
			key = identifier.ToNode()
			if hasPrecedingOriginalNameDirective {
				identifier.OriginalName = &originalNameDirectiveValue
			}
			token = p.lexer.Next()
		case DoubleQuoteString, SingleQuoteString:
			key = ast.NewStringLiteral(token.Value).ToNode()
			token = p.lexer.Next()
		case Decimal:
			key = ast.NewDecimalLiteral(token.Value).ToNode()
			token = p.lexer.Next()
		default:
			panic("Expected a valid key but got " + token.Value)
		}

		if token.Type == Colon {
			token = p.lexer.Next()
			body = append(body, ast.NewTPropertySignature(key, p.parseTTypeAnnotation()).ToNode())
			token = p.lexer.Peek()
			if token.Type == Comma {
				token = p.lexer.Next()
			}
		} else {
			if token.Type != Comma && token.Type != RightCurlyBrace {
				panic("A Expected '{' but got " + token.Value)
			}
			if token.Type == Comma {
				token = p.lexer.Next()
			}
			if key.Type == ast.NodeTypeIdentifier {
				body = append(body, ast.NewTPropertySignature(key, p.parseTTypeAnnotation()).ToNode())
			} else {
				panic("Expected Identifier but got " + token.Value)
			}
		}

		token = p.lexer.Peek()
	}
	if token.Type != RightCurlyBrace {
		panic("Expected '{' but got " + token.Value)
	}
	p.lexer.Next()

	token = p.lexer.Next()

	return ast.NewTInterfaceBody(body)
}

func (p *Parser) parseTTypeAnnotation() *ast.TTypeAnnotation {
	token := p.lexer.Peek()

	if token.Type == LeftParenthesis {
		return ast.NewTTypeAnnotation(p.parseTFunctionType().ToNode())
	}

	if token.Type == LeftCurlyBrace {
		return ast.NewTTypeAnnotation(p.parseTTypeLiteral().ToNode())
	}

	if token.Type != Identifier {
		panic("Expected Token Type Identifier, got " + token.Value)
	}
	p.lexer.Next()

	return ast.NewTTypeAnnotation(p.getTypeNodeFromIdentifier(token))
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	argument := p.parseExpression()
	if p.lexer.Peek().Type != Semicolon {
		panic("Expected '؛', got " + p.lexer.Peek().Value)
	}
	p.lexer.Next()

	return ast.NewReturnStatement(argument)
}

func (p *Parser) parseTypedIdentifier() *ast.Identifier {
	token := p.lexer.Peek()
	if token.Type != Identifier {
		panic("Expected Identifier, got " + token.Value)
	}
	identifierName := token.Value
	token = p.lexer.Next()

	var tTypeAnnotation *ast.TTypeAnnotation = nil
	if token.Type == Colon {
		token = p.lexer.Next()
		tTypeAnnotation = p.parseTTypeAnnotation()
	}

	return ast.NewIdentifier(identifierName, tTypeAnnotation)
}

func (p *Parser) parseTFunctionType() *ast.TFunctionType {
	token := p.lexer.Peek()
	if token.Type != LeftParenthesis {
		panic("Expected ')', got " + token.Value)
	}
	token = p.lexer.Next()
	params := []*ast.Identifier{}
	for token.Type != EOF && token.Type != Invalid && token.Type != RightParenthesis && token.Type == Identifier {
		params = append(params, p.parseTypedIdentifier())
		token = p.lexer.Peek()
		if token.Type != Comma && token.Type != RightParenthesis {
			panic("Expected ',', got " + token.Value)
		}
		if token.Type == Comma {
			token = p.lexer.Next()
		}
	}
	if token.Type != RightParenthesis {
		panic("Expected '(', got " + token.Value)
	}
	token = p.lexer.Next()

	if token.Type != EqualRightArrow {
		panic("Expected '=>', got " + token.Value)

	}
	p.lexer.Next()

	typeAnnotation := p.parseTTypeAnnotation()
	return ast.NewTFunctionType(params, typeAnnotation)
}

func (p *Parser) parseTTypeAliasDeclaration() *ast.TTypeAliasDeclaration {
	token := p.lexer.Peek()
	if token.Type != Identifier {
		panic("Expected Identifier, got " + token.Value)
	}
	identifier := ast.NewIdentifier(token.Value, nil)
	token = p.lexer.Next()

	if token.Type != Equal {
		panic("Expected '=', got " + token.Value)
	}
	token = p.lexer.Next()

	typeAnnotation := p.parseTTypeAnnotation()

	return ast.NewTTypeAliasDeclaration(identifier, typeAnnotation)
}

func (p *Parser) parseTTypeLiteral() *ast.TTypeLiteral {
	token := p.lexer.Peek()
	if token.Type != LeftCurlyBrace {
		panic("Expected '}' got " + token.Value)
	}
	token = p.lexer.Next()

	members := []*ast.Node{}
	for token.Type != EOF && token.Type != Invalid && token.Type != RightCurlyBrace {
		var key *ast.Node
		switch token.Type {
		case Identifier:
			key = ast.NewIdentifier(token.Value, nil).ToNode()
			token = p.lexer.Next()
		case DoubleQuoteString, SingleQuoteString:
			key = ast.NewStringLiteral(token.Value).ToNode()
			token = p.lexer.Next()
		case Decimal:
			key = ast.NewDecimalLiteral(token.Value).ToNode()
			token = p.lexer.Next()
		default:
			panic("Expected a valid key but got " + token.Value)
		}

		if token.Type == Colon {
			token = p.lexer.Next()
			members = append(members, ast.NewTPropertySignature(key, p.parseTTypeAnnotation()).ToNode())
			token = p.lexer.Peek()
			if token.Type == Comma {
				token = p.lexer.Next()
			}
		} else {
			if token.Type != Comma && token.Type != RightCurlyBrace {
				panic("A Expected '{' but got " + token.Value)
			}
			if token.Type == Comma {
				token = p.lexer.Next()
			}
			if key.Type == ast.NodeTypeIdentifier {
				members = append(members, ast.NewTPropertySignature(key, p.parseTTypeAnnotation()).ToNode())
			} else {
				panic("Expected Identifier but got " + token.Value)
			}
		}

		token = p.lexer.Peek()
	}
	if token.Type != RightCurlyBrace {
		panic("Expected '{' but got " + token.Value)
	}
	p.lexer.Next()

	token = p.lexer.Next()

	return ast.NewTTypeLiteral(members)
}

package lexer

import (
	"regexp"
	"unicode/utf8"
)

type Keyword = string

const (
	KeywordConst    Keyword = "ثابت"
	KeywordLet      Keyword = "متغير"
	KeywordFunction Keyword = "دالة"
	KeywordNull     Keyword = "فارغ"
	KeywordTrue     Keyword = "صحيح"
	KeywordFalse    Keyword = "خطأ"
	KeywordIf       Keyword = "إذا"
	KeywordElse     Keyword = "و_إلا"
	KeywordImport   Keyword = "استيراد"
	KeywordFrom     Keyword = "من"
	KeywordAs       Keyword = "بإسم"
	KeywordExport   Keyword = "تصدير"
	KeywordDefault  Keyword = "افتراضي"
	KeywordReturn   Keyword = "إرجاع"
	KeywordFor      Keyword = "من_أجل"
)

type TokenType int

const (
	KeywordToken TokenType = iota
	Identifier
	Decimal
	Equal
	Semicolon
	LeftSquareBracket
	RightSquareBracket
	LeftCurlyBrace
	RightCurlyBrace
	LeftParenthesis
	RightParenthesis
	EOF
	DoubleQuoteString
	SingleQuoteString
	Comma
	Dot
	Star
	BitwiseOr
	BitwiseXor
	BitwiseAnd
	EqualEqual
	NotEqual
	EqualEqualEqual
	NotEqualEqual
	LeftArrow
	RightArrow
	LeftArrowEqual
	RightArrowEqual
	DoubleLeftArrow
	DoubleRightArrow
	TripleRightArrow
	Plus
	Minus
	Slash
	Percent
	DoubleStar
	TripleDots
	Colon
	EqualRightArrow
	DoublePlus
	DoubleMinus
	StarEqual
	SlashEqual
	PercentEqual
	PlusEqual
	MinusEqual
	DoubleLeftArrowEqual
	DoubleRightArrowEqual
	BitwiseAndEqual
	BitwiseXorEqual
	BitwiseOrEqual
	DoubleStarEqual
	Invalid
)

func (t TokenType) String() string {
	switch t {
	case KeywordToken:
		return "KeywordToken"
	case Identifier:
		return "Identifier"
	case Decimal:
		return "Decimal"
	case Equal:
		return "Equal"
	case Semicolon:
		return "Semicolon"
	case LeftSquareBracket:
		return "LeftSquareBracket"
	case RightSquareBracket:
		return "RightSquareBracket"
	case LeftCurlyBrace:
		return "LeftCurlyBrace"
	case RightCurlyBrace:
		return "RightCurlyBrace"
	case LeftParenthesis:
		return "LeftParenthesis"
	case RightParenthesis:
		return "RightParenthesis"
	case EOF:
		return "EOF"
	case DoubleQuoteString:
		return "DoubleQuoteString"
	case SingleQuoteString:
		return "SingleQuoteString"
	case Comma:
		return "Comma"
	case Dot:
		return "Dot"
	case Star:
		return "Star"
	case BitwiseOr:
		return "BitwiseOr"
	case BitwiseXor:
		return "BitwiseXor"
	case BitwiseAnd:
		return "BitwiseAnd"
	case EqualEqual:
		return "EqualEqual"
	case NotEqual:
		return "NotEqual"
	case EqualEqualEqual:
		return "EqualEqualEqual"
	case NotEqualEqual:
		return "NotEqualEqual"
	case LeftArrow:
		return "LeftArrow"
	case RightArrow:
		return "RightArrow"
	case LeftArrowEqual:
		return "LeftArrowEqual"
	case RightArrowEqual:
		return "RightArrowEqual"
	case DoubleLeftArrow:
		return "DoubleLeftArrow"
	case DoubleRightArrow:
		return "DoubleRightArrow"
	case TripleRightArrow:
		return "TripleRightArrow"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Slash:
		return "Slash"
	case Percent:
		return "Percent"
	case DoubleStar:
		return "DoubleStar"
	case TripleDots:
		return "TripleDots"
	case Colon:
		return "Colon"
	case EqualRightArrow:
		return "EqualRightArrow"
	case DoublePlus:
		return "DoublePlus"
	case DoubleMinus:
		return "DoubleMinus"
	case StarEqual:
		return "StarEqual"
	case SlashEqual:
		return "SlashEqual"
	case PercentEqual:
		return "PercentEqual"
	case PlusEqual:
		return "PlusEqual"
	case MinusEqual:
		return "MinusEqual"
	case DoubleLeftArrowEqual:
		return "DoubleLeftArrowEqual"
	case DoubleRightArrowEqual:
		return "DoubleRightArrowEqual"
	case BitwiseAndEqual:
		return "BitwiseAndEqual"
	case BitwiseXorEqual:
		return "BitwiseXorEqual"
	case BitwiseOrEqual:
		return "BitwiseOrEqual"
	case DoubleStarEqual:
		return "DoubleStarEqual"
	case Invalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

type Token struct {
	Type     TokenType
	Value    string
	Position int
}

var Keywords = []Keyword{
	KeywordConst,
	KeywordLet,
	KeywordFunction,
	KeywordNull,
	KeywordTrue,
	KeywordFalse,
	KeywordIf,
	KeywordElse,
	KeywordImport,
	KeywordAs,
	KeywordExport,
	KeywordDefault,
	KeywordReturn,

	KeywordFor, // Must stay before KeywordFrom
	KeywordFrom,
}

var OneCharTokens = map[string]TokenType{
	"{": LeftCurlyBrace,
	"}": RightCurlyBrace,
	"(": LeftParenthesis,
	")": RightParenthesis,
	"[": LeftSquareBracket,
	"]": RightSquareBracket,
	"=": Equal,
	",": Comma,
	".": Dot,
	"*": Star,
	"|": BitwiseOr,
	"^": BitwiseXor,
	"&": BitwiseAnd,
	"<": LeftArrow,
	">": RightArrow,
	"+": Plus,
	"-": Minus,
	"/": Slash,
	"%": Percent,
	":": Colon,
}

var TwoCharTokens = map[string]TokenType{
	"==": EqualEqual,
	"!=": NotEqual,
	"<=": LeftArrowEqual,
	">=": RightArrowEqual,
	"<<": DoubleLeftArrow,
	">>": DoubleRightArrow,
	"**": DoubleStar,
	"=>": EqualRightArrow,
	"؛":  Semicolon, // Unicode character
	"++": DoublePlus,
	"--": DoubleMinus,
	"*=": StarEqual,
	"/=": SlashEqual,
	"%=": PercentEqual,
	"+=": PlusEqual,
	"-=": MinusEqual,
	"&=": BitwiseAndEqual,
	"^=": BitwiseXorEqual,
	"|=": BitwiseOrEqual,
}

var ThreeCharTokens = map[string]TokenType{
	"===": EqualEqualEqual,
	"!==": NotEqualEqual,
	">>>": TripleRightArrow,
	"...": TripleDots,
	"<<=": DoubleLeftArrowEqual,
	">>=": DoubleRightArrowEqual,
	"**=": DoubleStarEqual,
}

type TypeKeyword = string

const (
	TypeKeywordAny     TypeKeyword = "أي_نوع"
	TypeKeywordString  TypeKeyword = "نص"
	TypeKeywordNumber  TypeKeyword = "عدد"
	TypeKeywordBoolean TypeKeyword = "قيمة_منطقية"

	TypeKeywordInterface TypeKeyword = "واجهة"
	TypeKeywordType      TypeKeyword = "نوع"
	TypeKeywordDeclare   TypeKeyword = "تصريح"
	TypeKeywordModule    TypeKeyword = "وحدة"
)

type LexerState struct {
	position                          int
	startPosition                     int
	beforeWhitespacePosition          int
	currentToken                      Token
	HasPrecedingOriginalNameDirective bool
	OriginalNameDirectiveValue        string
}

type Lexer struct {
	input string
	LexerState
}

func NewLexer(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.position = 0
	lexer.HasPrecedingOriginalNameDirective = false
	lexer.currentToken = lexer.nextToken()
	return lexer
}

func (l *Lexer) Position() int                 { return l.position }
func (l *Lexer) StartPosition() int            { return l.startPosition }
func (l *Lexer) BeforeWhitespacePosition() int { return l.beforeWhitespacePosition }

func (l *Lexer) charAndSize() (rune, int) {
	return utf8.DecodeRuneInString(l.input[l.position:])
}

func (l *Lexer) current() string {
	if l.position >= len(l.input) {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.position:])
	return string(r)
}

func (l *Lexer) currentTwoChars() string {
	if l.position+1 >= len(l.input) {
		return ""
	}
	return l.input[l.position : l.position+2]
}

func (l *Lexer) currentThreeChars() string {
	if l.position+2 >= len(l.input) {
		return ""
	}
	return l.input[l.position : l.position+3]
}

func (l *Lexer) isEOF() bool {
	return l.position >= len(l.input)
}

func (l *Lexer) increasePosition(value int) {
	l.position += value
}

func (l *Lexer) isNewLineCharacter() bool {
	ch := l.current()
	return ch == "\r" || ch == "\n"
}

func (l *Lexer) isWhiteSpace() bool {
	ch := l.current()
	return ch == " " || ch == "\r" || ch == "\n" || ch == "\t"
}

func (l *Lexer) skipWhiteSpace() {
	for !l.isEOF() && l.isWhiteSpace() {
		l.increasePosition(1)
	}
}

func (l *Lexer) match(value string) bool {
	if l.isEOF() {
		return false
	}
	if l.position+len(value) > len(l.input) {
		return false
	}
	return l.input[l.position:l.position+len(value)] == value
}

func (l *Lexer) isArabicLetter() bool {
	ch := l.current()
	if ch == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(ch)
	return r >= 0x0620 && r <= 0x064A
}

func (l *Lexer) isDigit() bool {
	ch := l.current()
	return (ch >= "0" && ch <= "9") || (ch >= "٠" && ch <= "٩")
}

func (l *Lexer) Peek() Token {
	return l.currentToken
}

func (l *Lexer) Next() Token {
	l.HasPrecedingOriginalNameDirective = false
	l.OriginalNameDirectiveValue = ""
	l.beforeWhitespacePosition = l.position
	l.currentToken = l.nextToken()
	l.startPosition = l.currentToken.Position
	return l.currentToken
}

func (l *Lexer) nextToken() Token {
	l.skipWhiteSpace()

	if l.isEOF() {
		return Token{Type: EOF, Value: "", Position: l.position}
	}

	if l.currentTwoChars() == "//" {
		comment := "//"
		l.increasePosition(2)
		for !l.isEOF() && !l.isNewLineCharacter() {
			char, size := l.charAndSize()
			comment += string(char)
			l.increasePosition(size)
		}

		l.skipWhiteSpace()
		// Regex to extract value inside quotes
		re := regexp.MustCompile(`//\s*@الاسم_الأصلي\("([^"]+)"\)`)

		if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
			l.OriginalNameDirectiveValue = matches[1]
			l.HasPrecedingOriginalNameDirective = true
		}

		if l.isEOF() {
			return Token{Type: EOF, Value: l.current(), Position: l.position}
		}
	}

	for _, keyword := range Keywords {
		if l.match(string(keyword)) {
			pos := l.position
			l.increasePosition(len(keyword))
			return Token{Type: KeywordToken, Value: string(keyword), Position: pos}
		}
	}

	currentThreeChars := l.currentThreeChars()
	if tokenType, ok := ThreeCharTokens[currentThreeChars]; ok {
		pos := l.position
		l.increasePosition(3)
		return Token{Type: tokenType, Value: currentThreeChars, Position: pos}
	}

	currentTwoChars := l.currentTwoChars()
	if tokenType, ok := TwoCharTokens[currentTwoChars]; ok {
		pos := l.position
		l.increasePosition(2)
		return Token{Type: tokenType, Value: currentTwoChars, Position: pos}
	}

	currentOneChar := l.current()
	if tokenType, ok := OneCharTokens[currentOneChar]; ok {
		pos := l.position
		l.increasePosition(1)
		return Token{Type: tokenType, Value: currentOneChar, Position: pos}
	}

	if l.current() == `"` {
		start := l.position
		l.increasePosition(1)
		value := ""
		for !l.isEOF() && l.current() != `"` {
			char, size := l.charAndSize()
			value += string(char)
			l.increasePosition(size)
		}
		if l.isEOF() {
			return Token{Type: EOF, Value: l.current(), Position: l.position}
		}
		l.increasePosition(1)
		return Token{Type: DoubleQuoteString, Value: value, Position: start}
	}

	if l.current() == `'` {
		start := l.position
		l.increasePosition(1)
		value := ""
		for !l.isEOF() && l.current() != `'` {
			char, size := l.charAndSize()
			value += string(char)
			l.increasePosition(size)
		}
		if l.isEOF() {
			return Token{Type: EOF, Value: l.current(), Position: l.position}
		}
		l.increasePosition(1)
		return Token{Type: SingleQuoteString, Value: value, Position: start}
	}

	identifier := ""
	for !l.isEOF() && l.isArabicLetter() {
		char, size := l.charAndSize()
		identifier += string(char)
		l.increasePosition(size)
		if l.current() == "_" {
			identifier += "_"
			l.increasePosition(1)
		}
	}
	if len(identifier) > 0 {
		return Token{Type: Identifier, Value: identifier, Position: l.position - len(identifier)}
	}

	number := ""
	for !l.isEOF() && l.isDigit() {
		number += l.current()
		l.increasePosition(1)
	}
	if len(number) > 0 {
		return Token{Type: Decimal, Value: number, Position: l.position - len(number)}
	}

	return Token{Type: Invalid, Value: l.current(), Position: l.position}
}

func (l *Lexer) Mark() LexerState {
	return l.LexerState
}

func (l *Lexer) Rewind(state LexerState) {
	l.position = state.position
	l.startPosition = state.startPosition
	l.beforeWhitespacePosition = state.beforeWhitespacePosition
	l.currentToken = state.currentToken
	l.HasPrecedingOriginalNameDirective = state.HasPrecedingOriginalNameDirective
	l.OriginalNameDirectiveValue = state.OriginalNameDirectiveValue
}

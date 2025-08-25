package compiler

import (
	"unicode/utf8"
)

type Keyword = string
type TokenType = int

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
	Invalid
)

type Token struct {
	Type     TokenType
	Value    string
	Position int
}

var Keywords = []Keyword{
	"ثابت", "متغير", "دالة", "كائن", "فارغ", "صحيح", "خطأ", "إذا", "و_إلا",
	"استيراد", "من", "بإسم", "تصدير", "افتراضي",
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
	"؛":  Semicolon, // Unicode character
}

var ThreeCharTokens = map[string]TokenType{
	"===": EqualEqualEqual,
	"!==": NotEqualEqual,
	">>>": TripleRightArrow,
	"...": TripleDots,
}

type Lexer struct {
	input        string
	position     int
	currentToken Token
}

func NewLexer(input string) *Lexer {
	lexer := &Lexer{input: input, position: 0}
	lexer.currentToken = lexer.nextToken()
	return lexer
}

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
	l.currentToken = l.nextToken()
	return l.currentToken
}

func (l *Lexer) nextToken() Token {
	l.skipWhiteSpace()

	if l.isEOF() {
		return Token{Type: EOF, Value: "", Position: l.position}
	}

	for _, keyword := range Keywords {
		if l.match(string(keyword)) {
			pos := l.position
			l.increasePosition(len(keyword))
			return Token{Type: KeywordToken, Value: string(keyword), Position: pos}
		}
	}

	if tokenType, ok := ThreeCharTokens[l.currentThreeChars()]; ok {
		pos := l.position
		l.increasePosition(3)
		return Token{Type: tokenType, Value: l.currentThreeChars(), Position: pos}
	}

	if tokenType, ok := TwoCharTokens[l.currentTwoChars()]; ok {
		pos := l.position
		l.increasePosition(2)
		return Token{Type: tokenType, Value: l.currentTwoChars(), Position: pos}
	}

	if tokenType, ok := OneCharTokens[l.current()]; ok {
		pos := l.position
		l.increasePosition(1)
		return Token{Type: tokenType, Value: l.current(), Position: pos}
	}

	if l.current() == `"` {
		start := l.position
		l.increasePosition(1)
		value := `"`
		for !l.isEOF() && l.current() != `"` {
			char, size := l.charAndSize()
			value += string(char)
			l.increasePosition(size)
		}
		if l.isEOF() {
			return Token{Type: EOF, Value: l.current(), Position: l.position}
		}
		value += `"`
		l.increasePosition(1)
		return Token{Type: DoubleQuoteString, Value: value, Position: start}
	}

	if l.current() == `'` {
		start := l.position
		l.increasePosition(1)
		value := `'`
		for !l.isEOF() && l.current() != `'` {
			char, size := l.charAndSize()
			value += string(char)
			l.increasePosition(size)
		}
		if l.isEOF() {
			return Token{Type: EOF, Value: l.current(), Position: l.position}
		}
		value += `'`
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

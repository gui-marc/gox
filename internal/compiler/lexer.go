package compiler

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF       TokenType = iota
	TokenOpenTag             // <
	TokenCloseTag            // </
	TokenTagEnd              // >
	TokenSelfClose           // />
	TokenIdent               // tag or attribute name
	TokenString              // "string"
	TokenAssign              // =
	TokenText                // raw text outside tags
)

type Token struct {
	Type TokenType
	Val  string
	Pos  int
}

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

func (l *Lexer) Next() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF}
	}

	ch := l.input[l.pos]

	// Tag delimiters
	if ch == '<' {
		if l.match("</") {
			return Token{Type: TokenCloseTag, Val: "</", Pos: l.pos - 2}
		}
		l.pos++
		return Token{Type: TokenOpenTag, Val: "<", Pos: l.pos - 1}
	}

	if ch == '/' && l.peek() == '>' {
		l.pos += 2
		return Token{Type: TokenSelfClose, Val: "/>", Pos: l.pos - 2}
	}

	if ch == '>' {
		l.pos++
		return Token{Type: TokenTagEnd, Val: ">", Pos: l.pos - 1}
	}

	if ch == '=' {
		l.pos++
		return Token{Type: TokenAssign, Val: "=", Pos: l.pos - 1}
	}

	if ch == '"' {
		return l.scanString()
	}

	if isIdentStart(ch) {
		return l.scanIdent()
	}

	// Otherwise: text node
	return l.scanText()
}

// ------------------- helpers -------------------

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) match(s string) bool {
	for i := 0; i < len(s); i++ {
		if l.pos+i >= len(l.input) || rune(s[i]) != l.input[l.pos+i] {
			return false
		}
	}
	l.pos += len(s)
	return true
}

func (l *Lexer) scanString() Token {
	start := l.pos
	l.pos++ // skip opening quote
	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '"' {
			l.pos++
			return Token{Type: TokenString, Val: `"` + sb.String() + `"`, Pos: start}
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			// handle escaped characters
			sb.WriteRune(l.input[l.pos])
			l.pos++
			sb.WriteRune(l.input[l.pos])
			l.pos++
			continue
		}
		sb.WriteRune(ch)
		l.pos++
	}
	// Unterminated string
	return Token{Type: TokenString, Val: `"` + sb.String(), Pos: start}
}

func (l *Lexer) scanIdent() Token {
	start := l.pos
	for l.pos < len(l.input) && isIdentPart(l.input[l.pos]) {
		l.pos++
	}
	return Token{Type: TokenIdent, Val: string(l.input[start:l.pos]), Pos: start}
}

func (l *Lexer) scanText() Token {
	start := l.pos
	var sb strings.Builder
	for l.pos < len(l.input) && l.input[l.pos] != '<' {
		sb.WriteRune(l.input[l.pos])
		l.pos++
	}
	return Token{Type: TokenText, Val: sb.String(), Pos: start}
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == ':'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == ':' || ch == '-'
}

// ------------------- debug -------------------

func (t Token) String() string {
	return fmt.Sprintf("{Type: %v, Val: %q, Pos: %d}", t.Type, t.Val, t.Pos)
}

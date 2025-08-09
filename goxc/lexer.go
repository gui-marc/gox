// goxc/lexer.go
package main

import (
	"strings"
	"unicode"
)

type Lexer struct {
	source  []byte
	cursor  int  // current position in input (points to current char)
	readPos int  // current reading position in input (after current char)
	char    byte // current char under examination
}

func NewLexer(source []byte) *Lexer {
	l := &Lexer{source: source}
	l.readChar() // Initialize the first character
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.source) {
		l.char = 0 // NUL character signifies EOF
	} else {
		l.char = l.source[l.readPos]
	}
	l.cursor = l.readPos
	l.readPos++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.char {
	case '=':
		tok = newToken(EQUALS, l.char)
	case '/':
		tok = newToken(SLASH, l.char)
	case '<':
		// This is the only point where we need lookahead for the closing tag `</`
		if l.peekChar() == '/' {
			// This is a TEXT node start, not an element start.
			tok.Literal = l.readText()
			tok.Type = TEXT
			return tok
		}
		tok = newToken(LT, l.char)
	case '>':
		tok = newToken(GT, l.char)
	case '{':
		tok = newToken(LBRACE, l.char)
	case '}':
		tok = newToken(RBRACE, l.char)
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.char) {
			literal := l.readIdentifier()
			tok.Literal = literal
			tok.Type = IDENT
			return tok
		}
		// If it's not a special char and not an identifier, it must be a text node.
		tok.Literal = l.readText()
		tok.Type = TEXT
		return tok
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.char)) {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.cursor
	for isLetter(l.char) || unicode.IsDigit(rune(l.char)) {
		l.readChar()
	}
	return string(l.source[position:l.cursor])
}

func (l *Lexer) readString() string {
	position := l.cursor + 1 // Skip the opening "
	for {
		l.readChar()
		if l.char == '"' || l.char == 0 {
			break
		}
	}
	str := string(l.source[position:l.cursor])
	l.readChar() // Skip the closing "
	return str
}

func (l *Lexer) readText() string {
	position := l.cursor
	for l.char != '<' && l.char != '{' && l.char != 0 {
		l.readChar()
	}
	return strings.TrimSpace(string(l.source[position:l.cursor]))
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.source) {
		return 0
	}
	return l.source[l.readPos]
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

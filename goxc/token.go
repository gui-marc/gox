// goxc/token.go
package main

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL" // A token we don't know
	EOF     = "EOF"     // End of file

	// Identifiers & Literals
	IDENT  = "IDENT"  // div, MyComponent, class
	TEXT   = "TEXT"   // "Hello World"
	STRING = "STRING" // "container"

	// Operators & Delimiters
	EQUALS = "="
	SLASH  = "/" // /
	LT     = "<" // <
	GT     = ">" // >
	LBRACE = "{" // {
	RBRACE = "}" // }
)

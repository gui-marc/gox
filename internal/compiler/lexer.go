package compiler

import (
	"github.com/alecthomas/participle/v2/lexer"
)

func NewLexer() *lexer.StatefulDefinition {
	lexer := lexer.MustSimple([]lexer.SimpleRule{
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
		{"String", `"(\\"|[^"])*"`},
		{"Punct", `[-!$%^&*()_+|~=` + "`" + `{}\[\]:;'<>,.?/]`},
		{"Whitespace", `\s+`},
	})

	return lexer
}

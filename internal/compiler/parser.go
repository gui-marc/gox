package compiler

import (
	"github.com/alecthomas/participle/v2"
)

func NewParticipleParser() *participle.Parser[GOX] {
	lexer := NewLexer()

	parser := participle.MustBuild[GOX](
		participle.Lexer(lexer),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
	)

	return parser
}

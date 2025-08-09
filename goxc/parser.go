package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"strings"
	"unicode"
)

type Parser struct {
	source []byte
	out    bytes.Buffer
	s      scanner.Scanner
	fset   *token.FileSet

	// We need to keep track of the previous token to understand context.
	lastToken token.Token
}

func NewParser(source []byte) *Parser {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(source))

	var s scanner.Scanner
	s.Init(file, source, nil, scanner.ScanComments)

	return &Parser{
		source: source,
		s:      s,
		fset:   fset,
	}
}

func (p *Parser) Parse() ([]byte, error) {
	for {
		pos, tok, lit := p.s.Scan()
		if tok == token.EOF {
			break
		}

		// Heuristic 1: Detect multi-line, parenthesized GOX blocks.
		// e.g., return (<div...>)
		if tok == token.LPAREN && p.isGoxContext() {
			// Peek ahead to see if a '<' follows.
			nextPos, nextTok, _ := p.s.Scan()
			if nextTok == token.Token('<') {
				// It's a parenthesized GOX block.
				// We parse it but "swallow" the outer parentheses.
				goxCode, endPos, err := p.parseGoxElement(nextPos)
				if err != nil {
					return nil, err
				}
				p.out.Write(goxCode)

				// We must find the closing parenthesis and discard it.
				// Re-initialize scanner to continue after the block.
				p.reinitScanner(endPos)
				_, finalTok, _ := p.s.Scan()
				if finalTok != token.RPAREN {
					return nil, fmt.Errorf("expected ')' to close multi-line gox expression")
				}
				p.lastToken = token.RPAREN // Update state
				continue
			}
			// Not a GOX block, so put the token back and treat as normal.
			p.reinitScanner(p.fset.File(pos).Pos(int(pos)))
		}

		// Heuristic 2: Detect single-line GOX blocks.
		// e.g., return <div>...</div>
		if tok == token.Token('<') && p.isGoxContext() {
			// It's a single-line GOX block.
			goxCode, endPos, err := p.parseGoxElement(pos)
			if err != nil {
				return nil, err
			}
			p.out.Write(goxCode)

			// Re-initialize scanner to continue after the block.
			p.reinitScanner(endPos)
			p.lastToken = token.IDENT // Pretend we just wrote an identifier.
			continue
		}

		// Default: Not a GOX block, write the token as is.
		if lit != "" {
			p.out.WriteString(lit)
		} else {
			p.out.WriteString(tok.String())
		}
		p.out.WriteByte(' ')
		p.lastToken = tok
	}

	return p.out.Bytes(), nil
}

// reinitScanner moves the scanner to a new position in the source.
func (p *Parser) reinitScanner(offset token.Pos) {
	file := p.fset.AddFile("", p.fset.Base(), len(p.source))
	file.SetLinesForContent(p.source)
	p.s.Init(file, p.source[offset:], nil, scanner.ScanComments)
}

// isGoxContext checks if the previous token allows a GOX expression to start.
func (p *Parser) isGoxContext() bool {
	switch p.lastToken {
	case token.RETURN, token.ASSIGN, token.DEFINE, token.LPAREN, token.COMMA:
		return true
	default:
		return false
	}
}

// parseGoxElement is a manual, character-level parser for the <...> block.
// It returns the generated code, the new cursor position, and any error.
func (p *Parser) parseGoxElement(startOffset token.Pos) ([]byte, token.Pos, error) {
	subParser := &goxBlockParser{
		source: p.source,
		cursor: int(startOffset), // Start where the scanner left off
	}
	goxCode, err := subParser.parse()
	if err != nil {
		return nil, 0, err
	}
	return goxCode, token.Pos(subParser.cursor), nil
}

// --- GOX Block Parser (Manual Sub-Parser) ---
// This is very similar to our previous robust version.

type goxBlockParser struct {
	source []byte
	cursor int
	out    bytes.Buffer
}

func (p *goxBlockParser) parse() ([]byte, error) {
	if err := p.parseElement(); err != nil {
		return nil, err
	}

	return p.out.Bytes(), nil
}

func (p *goxBlockParser) parseElement() error {
	p.skipWhitespace()

	if p.source[p.cursor] != '<' {
		return fmt.Errorf("gox: expected '<' to start element")
	}

	p.cursor++ // consume '<'

	tagName := p.readIdentifier()
	isComponent := unicode.IsUpper(rune(tagName[0]))

	if isComponent {
		p.out.WriteString(tagName)
		p.out.WriteString("(runtime.Props{")
	} else {
		fmt.Fprintf(&p.out, `runtime.El("%s", runtime.Props{`, tagName)
	}

	p.skipWhitespace()
	for p.cursor < len(p.source) && p.source[p.cursor] != '>' && p.source[p.cursor] != '/' {
		key := p.readIdentifier()
		p.expect('=')

		fmt.Fprintf(&p.out, `"%s": `, key)

		if p.source[p.cursor] == '"' {
			p.out.WriteString(p.readStringLiteral())
		} else if p.source[p.cursor] == '{' {
			p.out.WriteString(p.readBraceContent())
		} else {
			return fmt.Errorf("gox: expected string or brace expression for attribute value")
		}

		p.out.WriteString(", ")

		p.skipWhitespace()
	}

	p.out.WriteString("}") // Close Props

	p.skipWhitespace()

	if p.cursor < len(p.source) && p.source[p.cursor] == '/' {
		p.expect('/')
		p.expect('>')
		p.out.WriteString(")")

		return nil
	}

	p.out.WriteString(", ")
	p.expect('>')

	hasWrittenFirstChild := false
	for {
		p.skipWhitespace()
		if p.cursor+1 < len(p.source) && p.source[p.cursor] == '<' && p.source[p.cursor+1] == '/' {
			break
		}

		if hasWrittenFirstChild {
			p.out.WriteString(", ")
		}

		if p.source[p.cursor] == '<' {
			if err := p.parseElement(); err != nil {
				return err
			}
			hasWrittenFirstChild = true
		} else if p.source[p.cursor] == '{' {
			content := p.readBraceContent()
			if strings.HasSuffix(strings.TrimSpace(content), "...") {
				p.out.WriteString(content)
			} else {
				fmt.Fprintf(&p.out, "runtime.Text(%s)", content)
			}
			hasWrittenFirstChild = true
		} else {
			text := p.readText()
			if strings.TrimSpace(text) != "" {
				fmt.Fprintf(&p.out, `runtime.Text("%s")`, text)
				hasWrittenFirstChild = true
			}
		}
	}

	p.expect('<')
	p.expect('/')
	closingTag := p.readIdentifier()
	if closingTag != tagName {
		return fmt.Errorf("gox: mismatched closing tag. Expected </%s> but got </%s>", tagName, closingTag)
	}
	p.expect('>')

	p.out.WriteString(")")

	return nil
}

// --- Lexer-like Helpers for sub-parser ---
func (p *goxBlockParser) skipWhitespace() {
	for p.cursor < len(p.source) && unicode.Is(unicode.Space, rune(p.source[p.cursor])) {
		p.cursor++
	}
}

func (p *goxBlockParser) expect(char byte) {
	p.skipWhitespace()
	if p.cursor < len(p.source) && p.source[p.cursor] == char {
		p.cursor++
	}
}

func (p *goxBlockParser) readIdentifier() string {
	p.skipWhitespace()
	start := p.cursor
	for p.cursor < len(p.source) && (unicode.IsLetter(rune(p.source[p.cursor])) || unicode.IsDigit(rune(p.source[p.cursor]))) {
		p.cursor++
	}
	return string(p.source[start:p.cursor])
}

func (p *goxBlockParser) readStringLiteral() string {
	p.expect('"')
	start := p.cursor

	for p.cursor < len(p.source) && p.source[p.cursor] != '"' {
		p.cursor++
	}

	str := string(p.source[start:p.cursor])

	p.expect('"')

	return fmt.Sprintf(`"%s"`, str)
}

func (p *goxBlockParser) readBraceContent() string {
	p.expect('{')
	start := p.cursor
	braceDepth := 1
	for p.cursor < len(p.source) && braceDepth > 0 {
		if p.source[p.cursor] == '{' {
			braceDepth++
		} else if p.source[p.cursor] == '}' {
			braceDepth--
		}
		p.cursor++
	}
	return string(p.source[start : p.cursor-1])
}

func (p *goxBlockParser) readText() string {
	start := p.cursor
	for p.cursor < len(p.source) && p.source[p.cursor] != '<' && p.source[p.cursor] != '{' {
		p.cursor++
	}
	return string(p.source[start:p.cursor])
}

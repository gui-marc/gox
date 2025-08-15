package compiler

import "fmt"

type Parser struct {
	lexer  *Lexer
	token  Token
	peeked bool
}

func NewParser(input string) *Parser {
	return &Parser{lexer: NewLexer(input)}
}

func (p *Parser) next() Token {
	if p.peeked {
		p.peeked = false
		return p.token
	}
	p.token = p.lexer.Next()
	return p.token
}

func (p *Parser) peek() Token {
	if !p.peeked {
		p.token = p.lexer.Next()
		p.peeked = true
	}
	return p.token
}

func (p *Parser) expect(tt TokenType) Token {
	tok := p.next()
	if tok.Type != tt {
		panic(fmt.Sprintf("expected token %v, got %v at pos %d", tt, tok.Type, tok.Pos))
	}
	return tok
}

func (p *Parser) Parse() *GOX {
	nodes := []*Node{}
	for p.peek().Type != TokenEOF {
		nodes = append(nodes, p.parseNode())
	}
	return &GOX{Nodes: nodes}
}

func (p *Parser) parseNode() *Node {
	p.expect(TokenOpenTag)
	tagTok := p.expect(TokenIdent)

	node := &Node{
		Tag:   tagTok.Val,
		Attrs: []*Attr{},
	}

	// parse attributes
	for p.peek().Type == TokenIdent {
		attr := &Attr{Name: p.next().Val}
		p.expect(TokenAssign)
		attr.Value = p.expect(TokenString).Val
		node.Attrs = append(node.Attrs, attr)
	}

	// parse self-closing or normal tag
	switch p.peek().Type {
	case TokenSelfClose:
		p.next()
		node.SelfClose = "/"
		return node
	case TokenTagEnd:
		p.next()
		node.SelfClose = ">"
	default:
		panic(fmt.Sprintf("unexpected token %v", p.peek()))
	}

	// parse children
	node.Children = []Node{}
	for {
		if p.peek().Type == TokenCloseTag {
			break
		}
		node.Children = append(node.Children, *p.parseNode())
	}

	// closing tag
	p.expect(TokenCloseTag)
	closeTok := p.expect(TokenIdent)
	if closeTok.Val != node.Tag {
		panic(fmt.Sprintf("mismatched closing tag: expected </%s>, got </%s>", node.Tag, closeTok.Val))
	}
	p.expect(TokenTagEnd)
	node.Close = &closeTok.Val

	return node
}

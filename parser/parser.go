package parser

import (
	"go9cc/token"
	"os"
)

const (
	ND_INFIX = "ND_INFIX"
	ND_NUM   = "ND_NUM"
)

type NodeKind string

type Node struct {
	Kind        NodeKind
	Left, Right *Node
	Val         int
	Str         string
	Token       *token.Token
}

func newNode(kind NodeKind, left, right *Node, val int, str string, token *token.Token) *Node {
	return &Node{Kind: kind, Left: left, Right: right, Val: val, Str: str, Token: token}
}

type Parser struct {
	tzer      *token.Tokenizer
	head, cur *token.Token
}

func New(tzer *token.Tokenizer) *Parser {
	parser := &Parser{tzer: tzer}
	parser.head = parser.tzer.Tokenize()
	parser.cur = parser.head
	return parser
}

func (p *Parser) Parse() *Node {
	return p.expr()
}

func (p *Parser) nextTkn() {
	if p.cur.Kind != token.EOF {
		p.cur = p.cur.Next
	}
}

func (p *Parser) expr() *Node {
	node := p.mul()

	for p.cur.Kind == token.PLUS || p.cur.Kind == token.MINUS {
		switch p.cur.Kind {
		case token.PLUS:
			fallthrough
		case token.MINUS:
			node = newNode(ND_INFIX, node, nil, p.cur.Val, p.cur.Str, p.cur)
			p.nextTkn()
			node.Right = p.mul()
		default:
			// never go here
			p.tzer.Error(p.cur.Col, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) mul() *Node {
	node := p.primary()

	for p.cur.Kind == token.ASTERISK || p.cur.Kind == token.SLASH {
		switch p.cur.Kind {
		case token.ASTERISK:
			fallthrough
		case token.SLASH:
			node = newNode(ND_INFIX, node, nil, p.cur.Val, p.cur.Str, p.cur)
			p.nextTkn()
			node.Right = p.primary()
		default:
			// never go here
			p.tzer.Error(p.cur.Col, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) primary() *Node {
	p.expect(p.cur, token.NUM, token.LPAREN)
	switch p.cur.Kind {
	case token.NUM:
		return p.num()
	case token.LPAREN:
		p.nextTkn() // (
		n := p.expr()
		p.expect(p.cur, token.RPAREN)
		p.nextTkn() // )
		return n
	default:
		// expectでチェックしているのでここは通らず.
		os.Exit(1)
		return nil
	}
}

func (p *Parser) num() *Node {
	p.expect(p.cur, token.NUM)
	node := newNode(ND_NUM, nil, nil, p.cur.Val, p.cur.Str, p.cur)
	p.nextTkn()
	return node
}

func (p *Parser) expect(token *token.Token, kinds ...token.TokenKind) {
	p.tzer.Expect(token, kinds...)
}

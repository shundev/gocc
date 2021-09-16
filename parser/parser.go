package parser

import (
	"bytes"
	"fmt"
	"go9cc/token"
	"os"
)

type Node interface {
	String() string
	Token() *token.Token
}

type TokenAccessor struct {
	token *token.Token
}

func (n *TokenAccessor) Token() *token.Token {
	return n.token
}

/* Num */

type NumNode struct {
	Val int
	TokenAccessor
}

func (n *NumNode) String() string {
	return fmt.Sprintf("%d", n.Val)
}

/* Infix */

type InfixNode struct {
	Left, Right Node
	Op          string
	TokenAccessor
}

func (n *InfixNode) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(n.Left.String())
	out.WriteString(" " + n.Op + " ")
	out.WriteString(n.Right.String())
	out.WriteString(")")
	return out.String()
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

func (p *Parser) Parse() Node {
	return p.expr()
}

func (p *Parser) nextTkn() {
	if p.cur.Kind != token.EOF {
		p.cur = p.cur.Next
	}
}

func (p *Parser) expr() Node {
	node := p.mul()

	for p.cur.Kind == token.PLUS || p.cur.Kind == token.MINUS {
		switch p.cur.Kind {
		case token.PLUS:
			fallthrough
		case token.MINUS:
			infix := &InfixNode{
				Left: node, Right: nil, Op: p.cur.Str,
				TokenAccessor: TokenAccessor{token: p.cur},
			}
			p.nextTkn()
			infix.Right = p.mul()
			node = infix
		default:
			// never go here
			p.tzer.Error(p.cur.Col, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) mul() Node {
	node := p.primary()

	for p.cur.Kind == token.ASTERISK || p.cur.Kind == token.SLASH {
		switch p.cur.Kind {
		case token.ASTERISK:
			fallthrough
		case token.SLASH:
			infix := &InfixNode{
				Left: node, Right: nil, Op: p.cur.Str,
				TokenAccessor: TokenAccessor{token: p.cur},
			}
			p.nextTkn()
			infix.Right = p.primary()
			node = infix
		default:
			// never go here
			p.tzer.Error(p.cur.Col, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) primary() Node {
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

func (p *Parser) num() Node {
	p.expect(p.cur, token.NUM)
	node := &NumNode{
		Val:           p.cur.Val,
		TokenAccessor: TokenAccessor{token: p.cur},
	}
	p.nextTkn()
	return node
}

func (p *Parser) expect(token *token.Token, kinds ...token.TokenKind) {
	p.tzer.Expect(token, kinds...)
}

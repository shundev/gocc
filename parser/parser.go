package parser

import "go9cc/token"

const (
	ND_ADD = "ND_ADD"
	ND_SUB = "ND_SUB"
	ND_MUL = "ND_MUL"
	ND_DIV = "ND_DIV"
	ND_NUM = "ND_NUM"
)

type NodeKind string

type Node struct {
	Kind        NodeKind
	Left, Right *Node
	Val         int
	Token       *token.Token
}

func newNode(kind NodeKind, left, right *Node, val int, token *token.Token) *Node {
	return &Node{Kind: kind, Left: left, Right: right, Val: val, Token: token}
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

func (p *Parser) expr() *Node {
	node := p.mul()

	for p.cur.Kind != token.EOF {
		p.Expect(p.cur, token.PLUS, token.MINUS)
	}

	return node
}

func (p *Parser) mul() *Node {

}

func (p *Parser) primary() *Node {

}

func (p *Parser) num() *Node {

}

func (p *Parser) expect(token *token.Token, kind token.TokenKind) {
	if token.Kind != kind {
		p.tzer.Error(token.Col, "Expected %s. Got %s.", kind, token.Kind)
	}
}

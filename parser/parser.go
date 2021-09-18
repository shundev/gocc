package parser

import (
	"bytes"
	"fmt"
	"go9cc/token"
	"os"
	"strings"
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

/* Unary */

type UnaryNode struct {
	Right Node
	Op    string
	TokenAccessor
}

func (n *UnaryNode) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(n.Op)
	out.WriteString(n.Right.String())
	out.WriteString(")")
	return out.String()
}

/* Identifier */

type IdentNode struct {
	Name string
	TokenAccessor
}

func (n *IdentNode) String() string {
	return n.Name
}

/* Statement */

type StmtNode struct {
	Exp Node
	TokenAccessor
}

func (n *StmtNode) String() string {
	return n.Exp.String()
}

/* Program */

type ProgramNode struct {
	Stmts []*StmtNode
	TokenAccessor
}

func (n *ProgramNode) String() string {
	ss := []string{}
	for _, stmt := range n.Stmts {
		ss = append(ss, stmt.String())
	}
	return strings.Join(ss, "; ")
}

/*
expr    = assign
assign  = eq ("=" assign)?
eq      = lg ("==" lg)?
lg      = add ("<" add)?
add     = mul ("+" mul | "-" mul)*
mul     = unary ("*" unary | "/" unary)*
unary   = ("+" | "-")? primary
primary = num | ident | "(" expr ")"
*/

/* Parser */

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
	return p.program()
}

func UnaryToInfix(unary *UnaryNode) Node {
	right := unary.Right
	left := &NumNode{Val: 0}
	infix := &InfixNode{Left: left, Right: right, Op: unary.Op}
	return infix
}

func Swap(infix *InfixNode) Node {
	right := infix.Right
	left := infix.Left
	infix.Right = left
	infix.Left = right
	return infix
}

func (p *Parser) nextTkn() {
	if p.cur.Kind != token.EOF {
		p.cur = p.cur.Next
	}
}

func (p *Parser) program() Node {
	node := &ProgramNode{}
	node.Stmts = []*StmtNode{}
	for p.cur.Kind != token.EOF {
		node.Stmts = append(node.Stmts, p.stmt())
		if p.cur.Kind != token.SEMICOLLON {
			break
		}

		p.nextTkn()
	}
	return node
}

func (p *Parser) stmt() *StmtNode {
	exp := p.expr()
	node := &StmtNode{Exp: exp}
	return node
}

func (p *Parser) expr() Node {
	return p.assign()
}

func (p *Parser) assign() Node {
	node := p.eq()

	if p.cur.Kind == token.ASSIGN {
		infix := &InfixNode{
			Left: node, Right: nil, Op: p.cur.Str,
			TokenAccessor: TokenAccessor{token: p.cur},
		}
		p.nextTkn() // =
		infix.Right = p.assign()
		node = infix
	}

	return node
}

func (p *Parser) eq() Node {
	node := p.lg()

	if p.cur.Kind == token.EQ || p.cur.Kind == token.NEQ {
		infix := &InfixNode{
			Left: node, Right: nil, Op: p.cur.Str,
			TokenAccessor: TokenAccessor{token: p.cur},
		}
		p.nextTkn()
		infix.Right = p.lg()
		node = infix
	}

	return node
}

func (p *Parser) lg() Node {
	node := p.add()

	switch p.cur.Kind {
	case token.LT:
		fallthrough
	case token.GT:
		fallthrough
	case token.LTE:
		fallthrough
	case token.GTE:
		infix := &InfixNode{
			Left: node, Right: nil, Op: p.cur.Str,
			TokenAccessor: TokenAccessor{token: p.cur},
		}
		p.nextTkn()
		infix.Right = p.add()
		node = infix
	}

	return node
}

func (p *Parser) add() Node {
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
	node := p.unary()

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
			infix.Right = p.unary()
			node = infix
		default:
			// never go here
			p.tzer.Error(p.cur.Col, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) unary() Node {
	switch p.cur.Kind {
	case token.PLUS:
		fallthrough
	case token.MINUS:
		node := &UnaryNode{
			Right:         nil,
			Op:            p.cur.Str,
			TokenAccessor: TokenAccessor{token: p.cur},
		}
		p.nextTkn()
		node.Right = p.primary()
		return node
	default:
		n := p.primary()
		return n
	}
}

func (p *Parser) primary() Node {
	p.expect(p.cur, token.NUM, token.IDENT, token.LPAREN)
	switch p.cur.Kind {
	case token.NUM:
		return p.num()
	case token.IDENT:
		return p.ident()
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

func (p *Parser) ident() Node {
	p.expect(p.cur, token.IDENT)
	node := &IdentNode{
		Name:          p.cur.Str,
		TokenAccessor: TokenAccessor{token: p.cur},
	}
	p.nextTkn()
	return node
}

func (p *Parser) expect(token *token.Token, kinds ...token.TokenKind) {
	p.tzer.Expect(token, kinds...)
}

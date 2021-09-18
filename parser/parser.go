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
	TokenLiteral() string
}

type Stmt interface {
	Node
	stmtNode()
}

type Exp interface {
	Node
	expNode()
}

/* Num */

type NumExp struct {
	Val   int
	token *token.Token
}

func (n *NumExp) expNode() {}

func (n *NumExp) TokenLiteral() string {
	return n.token.Str
}

func (n *NumExp) String() string {
	return fmt.Sprintf("%d", n.Val)
}

/* Infix */

type InfixExp struct {
	Left, Right Exp
	Op          string
	token       *token.Token
}

func (n *InfixExp) expNode() {}

func (n *InfixExp) TokenLiteral() string {
	return n.token.Str
}

func (n *InfixExp) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(n.Left.String())
	out.WriteString(" " + n.Op + " ")
	out.WriteString(n.Right.String())
	out.WriteString(")")
	return out.String()
}

/* Unary */

type UnaryExp struct {
	Right Exp
	Op    string
	token *token.Token
}

func (n *UnaryExp) expNode() {}

func (n *UnaryExp) TokenLiteral() string {
	return n.token.Str
}

func (n *UnaryExp) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(n.Op)
	out.WriteString(n.Right.String())
	out.WriteString(")")
	return out.String()
}

/* Identifier */

type IdentExp struct {
	Name  string
	token *token.Token
}

func (n *IdentExp) expNode() {}

func (n *IdentExp) TokenLiteral() string {
	return n.token.Str
}

func (n *IdentExp) String() string {
	return n.Name
}

/* Statement */

type ExpStmt struct {
	Exp   Exp
	token *token.Token
}

func (n *ExpStmt) stmtNode() {}

func (n *ExpStmt) TokenLiteral() string {
	return n.token.Str
}

func (n *ExpStmt) String() string {
	var out bytes.Buffer
	out.WriteString(n.Exp.String())
	out.WriteString(";")
	return out.String()
}

/* Return Statement */

type ReturnStmt struct {
	Exp   Exp
	token *token.Token
}

func (n *ReturnStmt) stmtNode() {}

func (n *ReturnStmt) TokenLiteral() string {
	return n.token.Str
}

func (n *ReturnStmt) String() string {
	var out bytes.Buffer
	out.WriteString("return ")
	out.WriteString(n.Exp.String())
	out.WriteString(";")
	return out.String()
}

/* If Statement */

type IfStmt struct {
	Cond     Exp
	IfBody   Stmt
	ElseBody Stmt
	token    *token.Token
}

func (n *IfStmt) stmtNode() {}

func (n *IfStmt) TokenLiteral() string {
	return n.token.Str
}

func (n *IfStmt) String() string {
	var out bytes.Buffer
	out.WriteString("if (")
	out.WriteString(n.Cond.String())
	out.WriteString(") { ")
	out.WriteString(n.IfBody.String())
	out.WriteString(" }")
	if n.ElseBody != nil {
		out.WriteString(" else { ")
		out.WriteString(n.ElseBody.String())
		out.WriteString(" }")
	}
	return out.String()
}

/* While Statement */

type WhileStmt struct {
	Cond  Exp
	Body  Stmt
	token *token.Token
}

func (n *WhileStmt) stmtNode() {}

func (n *WhileStmt) TokenLiteral() string {
	return n.token.Str
}

func (n *WhileStmt) String() string {
	var out bytes.Buffer
	out.WriteString("while (")
	out.WriteString(n.Cond.String())
	out.WriteString(") { ")
	out.WriteString(n.Body.String())
	out.WriteString(" }")
	return out.String()
}

/* Program */

type ProgramNode struct {
	Stmts []Stmt
}

func (n *ProgramNode) TokenLiteral() string {
	if len(n.Stmts) > 0 {
		return n.Stmts[0].TokenLiteral()
	}

	return ""
}

func (n *ProgramNode) String() string {
	ss := []string{}
	for _, stmt := range n.Stmts {
		ss = append(ss, stmt.String())
	}
	return strings.Join(ss, " ")
}

/*
program   = stmt*
stmt      = (return expr ";") | (expr ";") | ifstmt | whilestmt
ifstmt    = "if" "(" expr ")" stmt ("else" stmt)?
whilestmt = "while" "(" expr ")" stmt
expr      = assign
assign    = eq ("=" assign)?
eq        = lg ("==" lg)?
lg        = add ("<" add)?
add       = mul ("+" mul | "-" mul)*
mul       = unary ("*" unary | "/" unary)*
unary     = ("+" | "-")? primary
primary   = num | ident | "(" expr ")"
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

func (p *Parser) Parse() *ProgramNode {
	return p.program()
}

func UnaryToInfix(unary *UnaryExp) *InfixExp {
	right := unary.Right
	left := &NumExp{Val: 0}
	infix := &InfixExp{Left: left, Right: right, Op: unary.Op}
	return infix
}

func Swap(infix *InfixExp) *InfixExp {
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

func (p *Parser) program() *ProgramNode {
	node := &ProgramNode{}
	node.Stmts = []Stmt{}
	for p.cur.Kind != token.EOF {
		node.Stmts = append(node.Stmts, p.stmt())
	}
	return node
}

func (p *Parser) stmt() Stmt {
	if p.cur.Kind == token.WHILE {
		return p.whileStmt()
	}

	if p.cur.Kind == token.IF {
		return p.ifStmt()
	}

	if p.cur.Kind == token.RETURN {
		return p.returnStmt()
	}

	exp := p.expr()
	node := &ExpStmt{Exp: exp}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()
	return node
}

func (p *Parser) whileStmt() *WhileStmt {
	p.expect(p.cur, token.WHILE)
	tkn := p.cur
	p.nextTkn()
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()
	exp := p.expr()
	p.expect(p.cur, token.RPAREN)
	p.nextTkn()
	body := p.stmt()
	node := &WhileStmt{
		Cond:  exp,
		Body:  body,
		token: tkn,
	}
	return node
}

func (p *Parser) ifStmt() *IfStmt {
	p.expect(p.cur, token.IF)
	tkn := p.cur
	p.nextTkn()
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()
	exp := p.expr()
	p.expect(p.cur, token.RPAREN)
	p.nextTkn()
	ifBody := p.stmt()
	node := &IfStmt{
		Cond:   exp,
		IfBody: ifBody,
		token:  tkn,
	}
	if p.cur.Kind == token.ELSE {
		p.nextTkn()
		node.ElseBody = p.stmt()
	}
	return node
}

func (p *Parser) returnStmt() *ReturnStmt {
	p.expect(p.cur, token.RETURN)
	tkn := p.cur
	p.nextTkn()
	exp := p.expr()
	node := &ReturnStmt{
		Exp:   exp,
		token: tkn,
	}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()
	return node
}

func (p *Parser) expr() Exp {
	return p.assign()
}

func (p *Parser) assign() Exp {
	node := p.eq()

	if p.cur.Kind == token.ASSIGN {
		infix := &InfixExp{
			Left: node, Right: nil, Op: p.cur.Str, token: p.cur,
		}
		p.nextTkn() // =
		infix.Right = p.assign()
		node = infix
	}

	return node
}

func (p *Parser) eq() Exp {
	node := p.lg()

	if p.cur.Kind == token.EQ || p.cur.Kind == token.NEQ {
		infix := &InfixExp{
			Left: node, Right: nil, Op: p.cur.Str, token: p.cur,
		}
		p.nextTkn()
		infix.Right = p.lg()
		node = infix
	}

	return node
}

func (p *Parser) lg() Exp {
	node := p.add()

	switch p.cur.Kind {
	case token.LT:
		fallthrough
	case token.GT:
		fallthrough
	case token.LTE:
		fallthrough
	case token.GTE:
		infix := &InfixExp{
			Left: node, Right: nil, Op: p.cur.Str, token: p.cur,
		}
		p.nextTkn()
		infix.Right = p.add()
		node = infix
	}

	return node
}

func (p *Parser) add() Exp {
	node := p.mul()

	for p.cur.Kind == token.PLUS || p.cur.Kind == token.MINUS {
		switch p.cur.Kind {
		case token.PLUS:
			fallthrough
		case token.MINUS:
			infix := &InfixExp{
				Left: node, Right: nil, Op: p.cur.Str, token: p.cur,
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

func (p *Parser) mul() Exp {
	node := p.unary()

	for p.cur.Kind == token.ASTERISK || p.cur.Kind == token.SLASH {
		switch p.cur.Kind {
		case token.ASTERISK:
			fallthrough
		case token.SLASH:
			infix := &InfixExp{
				Left: node, Right: nil, Op: p.cur.Str, token: p.cur,
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

func (p *Parser) unary() Exp {
	switch p.cur.Kind {
	case token.PLUS:
		fallthrough
	case token.MINUS:
		node := &UnaryExp{
			Right: nil,
			Op:    p.cur.Str,
			token: p.cur,
		}
		p.nextTkn()
		node.Right = p.primary()
		return node
	default:
		n := p.primary()
		return n
	}
}

func (p *Parser) primary() Exp {
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

func (p *Parser) num() Exp {
	p.expect(p.cur, token.NUM)
	node := &NumExp{
		Val: p.cur.Val, token: p.cur,
	}
	p.nextTkn()
	return node
}

func (p *Parser) ident() Exp {
	p.expect(p.cur, token.IDENT)
	node := &IdentExp{
		Name: p.cur.Str, token: p.cur,
	}
	p.nextTkn()
	return node
}

func (p *Parser) expect(token *token.Token, kinds ...token.TokenKind) {
	p.tzer.Expect(token, kinds...)
}

package parser

import (
	"bytes"
	"fmt"
	"go9cc/token"
	"go9cc/types"
	"os"
	"strings"
)

type LocalVariable struct {
	Name   string
	Type   types.Type
	offset int
}

func (n *LocalVariable) String() string {
	return n.Name
}

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
	Type() types.Type
}

/* Local Variable */

type LocalVariableNode struct {
	Locals []*LocalVariable
	Type   types.Type
	token  *token.Token
}

func (n *LocalVariableNode) TokenLiteral() string {
	return n.token.Str
}

func (n *LocalVariableNode) String() string {
	var out bytes.Buffer
	out.WriteString(n.Type.String())
	out.WriteString(" ")

	ss := []string{}
	for _, local := range n.Locals {
		ss = append(ss, local.String())
	}
	out.WriteString(strings.Join(ss, ", "))
	return out.String()
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

func (n *NumExp) Type() types.Type {
	return types.GetInt()
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

func (n *InfixExp) Type() types.Type {
	return n.Right.Type()
}

/* Declaration */

type DeclarationExp struct {
	LV    *LocalVariableNode
	Exp   Exp
	Op    string
	token *token.Token
}

func (n *DeclarationExp) expNode() {}

func (n *DeclarationExp) TokenLiteral() string {
	return n.token.Str
}

func (n *DeclarationExp) String() string {
	var out bytes.Buffer
	out.WriteString(n.LV.String())
	if n.Exp != nil {
		out.WriteString(" " + n.Op + " ")
		out.WriteString(n.Exp.String())
	}
	return out.String()
}

func (n *DeclarationExp) Type() types.Type {
	return n.LV.Type
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

func (n *UnaryExp) Type() types.Type {
	switch n.Op {
	case "+":
		fallthrough
	case "-":
		fallthrough
	case "*":
		return types.GetInt()
	case "&":
		return types.PointerTo(types.GetInt())
	}

	fmt.Fprintf(os.Stderr, "Invalid op: %s", n.Op)
	os.Exit(1)
	return types.GetInt()
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

func (n *IdentExp) Type() types.Type {
	return types.GetInt()
}

/* Function */

type FuncCallExp struct {
	Name  string
	token *token.Token
}

func (n *FuncCallExp) expNode() {}

func (n *FuncCallExp) TokenLiteral() string {
	return n.token.Str
}

func (n *FuncCallExp) Type() types.Type {
	// FIXME
	return types.GetInt()
}

func (n *FuncCallExp) String() string {
	var out bytes.Buffer
	out.WriteString(n.Name)
	out.WriteString(" (")
	out.WriteString(")")
	return out.String()
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
	out.WriteString(") ")
	out.WriteString(n.IfBody.String())
	if n.ElseBody != nil {
		out.WriteString(" else ")
		out.WriteString(n.ElseBody.String())
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
	out.WriteString(") ")
	out.WriteString(n.Body.String())
	return out.String()
}

/* For Statement */

type ForStmt struct {
	Init, Cond, AfterEach Exp
	Body                  Stmt
	token                 *token.Token
}

func (n *ForStmt) stmtNode() {}

func (n *ForStmt) TokenLiteral() string {
	return n.token.Str
}

func (n *ForStmt) String() string {
	var out bytes.Buffer
	out.WriteString("for (")
	if n.Init != nil {
		out.WriteString(n.Init.String())
	}
	out.WriteString(";")
	if n.Cond != nil {
		out.WriteString(n.Cond.String())
	}
	out.WriteString(";")
	if n.AfterEach != nil {
		out.WriteString(n.AfterEach.String())
	}
	out.WriteString(") ")
	out.WriteString(n.Body.String())
	return out.String()
}

/* Block Statement */

type BlockStmt struct {
	Stmts []Stmt
	token *token.Token
}

func (n *BlockStmt) stmtNode() {}

func (n *BlockStmt) TokenLiteral() string {
	return n.token.Str
}

func (n *BlockStmt) String() string {
	var out bytes.Buffer
	ss := []string{}
	for _, stmt := range n.Stmts {
		ss = append(ss, stmt.String())
	}

	out.WriteString("{ ")
	out.WriteString(strings.Join(ss, " "))
	out.WriteString(" }")
	return out.String()
}

/* Program */

type ProgramNode struct {
	Stmts   []Stmt
	Offsets map[string]int
}

func (n *ProgramNode) StackSize() int {
	max := 0
	for _, v := range n.Offsets {
		if v > max {
			max = v
		}
	}

	return alignTo(max, 16)
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
stmt      = (declaration ";") | (return expr ";") | (expr ";") | ifstmt | whilestmt | blockstmt
blockstmt = "{" stmt* "}"
forstmt   = "for" "(" (expr|declaration)? ";" expr? ";" expr? ")" stmt
ifstmt    = "if" "(" expr ")" stmt ("else" stmt)?
whilestmt = "while" "(" expr ")" stmt
expr      = assign
assign    = eq ("=" assign)?
eq        = lg ("==" lg)?
lg        = add ("<" add)?
add       = mul ("+" mul | "-" mul)*
mul       = unary ("*" unary | "/" unary)*
unary     = ("+" | "-")? primary
primary   = num | funccall | ident | "(" expr ")"
funccall  = ident "(" ")"

declaration =
  declspec
    (declarator
      ("=" expr)?
      ("," declarator ("=" expr)?)
    *)?
  ";"
declarator = "*"* ident
declspec = "int"
*/

/* Parser */

type Parser struct {
	tzer      *token.Tokenizer
	head, cur *token.Token
	offsetCnt int
	offsets   map[string]int
}

func New(tzer *token.Tokenizer) *Parser {
	parser := &Parser{tzer: tzer}
	parser.head = parser.tzer.Tokenize()
	parser.cur = parser.head
	parser.offsets = make(map[string]int)
	return parser
}

func (p *Parser) Parse() *ProgramNode {
	node := p.program()
	node.Offsets = p.offsets
	return node
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
	if p.cur.Kind == token.TYPE {
		return p.declarationStmt()
	}

	if p.cur.Kind == token.LBRACE {
		return p.blockStmt()
	}

	if p.cur.Kind == token.FOR {
		return p.forStmt()
	}

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

func (p *Parser) declspec() types.Type {
	p.expect(p.cur, token.TYPE)
	p.nextTkn()
	return types.GetInt()
}

// declarator = "*"* ident
func (p *Parser) declarator(ty types.Type) (types.Type, *token.Token) {
	for p.cur.Kind == token.ASTERISK {
		ty = types.PointerTo(ty)
		p.nextTkn()
	}

	p.expect(p.cur, token.IDENT)
	switch ty := ty.(type) {
	case *types.Int:
		return ty, p.cur
	case *types.IntPointer:
		return ty, p.cur
	default:
		fmt.Fprintf(os.Stderr, "Invalid type.")
		os.Exit(1)
	}

	return ty, nil
}

func (p *Parser) declaration() *DeclarationExp {
	initTok := p.cur
	baseTy := p.declspec()               // "int"
	ty, identTok := p.declarator(baseTy) // "**a"
	p.nextTkn()
	p.expect(p.cur, token.ASSIGN)
	local := &LocalVariable{Name: identTok.Str, Type: ty}
	left := &LocalVariableNode{Locals: []*LocalVariable{local}, Type: ty, token: initTok}
	p.nextTkn() // "="
	right := p.expr()
	dec := &DeclarationExp{LV: left, Exp: right, Op: "=", token: initTok}
	for _, local := range dec.LV.Locals {
		if _, exists := p.offsets[local.Name]; exists {
			p.tzer.Error(p.cur.Col, "Redefinition found: %s", p.cur.Str)
		}

		p.offsetCnt += 8
		p.offsets[local.Name] = p.offsetCnt
	}
	return dec
}

func (p *Parser) declarationStmt() *ExpStmt {
	dec := p.declaration()
	node := &ExpStmt{Exp: dec, token: dec.token}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()

	return node
}

func (p *Parser) blockStmt() *BlockStmt {
	p.expect(p.cur, token.LBRACE)
	tkn := p.cur
	p.nextTkn() // {
	node := &BlockStmt{token: tkn}
	node.Stmts = []Stmt{}
	for p.cur.Kind != token.RBRACE {
		node.Stmts = append(node.Stmts, p.stmt())
	}
	p.nextTkn() // }
	return node
}

func (p *Parser) forStmt() *ForStmt {
	p.expect(p.cur, token.FOR)
	tkn := p.cur
	node := &ForStmt{token: tkn}
	p.nextTkn()
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()

	if p.cur.Kind != token.SEMICOLLON {
		if p.cur.Kind == token.TYPE {
			node.Init = p.declaration()
		} else {
			node.Init = p.expr()
		}
	}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()

	if p.cur.Kind != token.SEMICOLLON {
		node.Cond = p.expr()
	}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()

	if p.cur.Kind != token.RPAREN {
		node.AfterEach = p.expr()
	}
	p.expect(p.cur, token.RPAREN)
	p.nextTkn()

	node.Body = p.stmt()
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

		// TODO: duplicate left value check
		if ident, ok := infix.Left.(*IdentExp); ok {
			if _, exists := p.offsets[ident.Name]; !exists {
				p.tzer.Error(ident.token.Col, "Variable not found.")
			}
		}
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
		fallthrough
	case token.ASTERISK:
		fallthrough
	case token.AND:
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
	case token.AND:
		fallthrough
	case token.ASTERISK:
		fallthrough
	case token.PLUS:
		fallthrough
	case token.MINUS:
		return p.unary()
	default:
		p.tzer.Error(p.cur.Col, "Invalid token as primary: %s", p.cur.Str)
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
	tkn := p.cur
	p.nextTkn()

	if p.cur.Kind == token.LPAREN {
		p.nextTkn()
		p.expect(p.cur, token.RPAREN)
		p.nextTkn()
		return &FuncCallExp{
			Name: tkn.Str, token: tkn,
		}
	} else {
		return &IdentExp{
			Name: tkn.Str, token: tkn,
		}
	}
}

func Scale(infix *InfixExp) *InfixExp {
	right, ok := infix.Right.(*NumExp)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to scale.")
		os.Exit(1)
	}

	num8 := &NumExp{Val: 8}
	mul := &InfixExp{Left: right, Right: num8, Op: "*"}
	infix.Right = mul
	return infix
}

func (p *Parser) expect(token *token.Token, kinds ...token.TokenKind) {
	p.tzer.Expect(token, kinds...)
}

func alignTo(n, align int) int {
	return (n + align - 1) / align * align
}

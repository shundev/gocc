package parser

import (
	"fmt"
	"go9cc/ast"
	"go9cc/token"
	"go9cc/types"
	"os"
	"strconv"
)

const DEBUG = true

/*
program     = (funcdef | global)*
global      = declaration
funcdef     = declspec declarator funcargs blockStmt
funcargs    = "(" declspec declarator ("," declspec declarator)* ")" | "(" ")"
blockstmt   = "{" stmt* "}"
stmt        = (declaration ";") | (return expr ";") | (expr ";") | ifstmt | whilestmt | blockstmt
forstmt     = "for" "(" (expr|declaration)? ";" expr? ";" expr? ")" stmt
ifstmt      = "if" "(" expr ")" stmt ("else" stmt)?
whilestmt   = "while" "(" expr ")" stmt
expr        = assign
assign      = eq ("=" assign)?
eq          = lg ("==" lg)?
lg          = add ("<" add)?
add         = mul ("+" mul | "-" mul)*
mul         = unary ("*" unary | "/" unary)*
unary       = ("+" | "-" | "sizeof")? primary
primary     = (ident "[" expr "]") | string | num | funccall | ident | "(" expr ")"
funccall    = ident funcparams
funcparams  = "(" ( expr ("," expr)* ")" | ")")

declaration =
  declspec
    (declarator
      ("=" expr)?
      ("," declarator ("=" expr)?)
    *)?
  ";"
declarator = "*"* ident ("[" num "]")?
declspec = "int"
*/

/* Parser */

type Parser struct {
	tzer      *token.Tokenizer
	head, cur *token.Token
	curFn     *ast.FuncDefNode
	Globals   map[string]*ast.LocalVariable
	funcdefs  map[string]*ast.FuncDefNode
	Strings   []*ast.StringLiteralExp
	strCnt    int
}

func New(tzer *token.Tokenizer) *Parser {
	parser := &Parser{
		tzer:     tzer,
		Globals:  map[string]*ast.LocalVariable{},
		funcdefs: map[string]*ast.FuncDefNode{},
		Strings:  []*ast.StringLiteralExp{},
	}
	parser.head = parser.tzer.Tokenize()
	parser.cur = parser.head
	return parser
}

func (p *Parser) Parse() *ast.ProgramNode {
	node := p.program()
	for _, funcdef := range node.FuncDefs {
		funcdef.PrepareStackSize()
	}
	return node
}

func (p *Parser) Error(token *token.Token, msg string, args ...interface{}) {
	p.tzer.Error(token, msg, args...)
}

func (p *Parser) nextTkn() {
	if p.cur.Kind != token.EOF {
		p.cur = p.cur.Next
	}
}

func (p *Parser) backTo(to *token.Token) {
	for p.cur != to {
		p.cur = p.cur.Prev
	}
}

func (p *Parser) getDef(name string) *ast.LocalVariable {
	debug("getDef def not found: %s", name)
	debug("getDef p.Globals: %s", p.Globals)
	debug("getDef p.curFn: %s", p.curFn)
	if p.curFn != nil {
		debug("getDef p.Locals: %s", p.curFn.Locals)
		if v, ok := p.curFn.Locals[name]; ok {
			return v
		}
	}

	if v, ok := p.Globals[name]; ok {
		return v
	}

	err("Ident %s not defined.\n", name)
	os.Exit(1)
	return nil
}

func (p *Parser) program() *ast.ProgramNode {
	node := &ast.ProgramNode{}
	node.FuncDefs = []*ast.FuncDefNode{}
	node.GlobalStmts = []*ast.DeclarationStmt{}
	for p.cur.Kind != token.EOF {
		n := p.global()
		switch n := n.(type) {
		case *ast.StmtListNode:
			for _, stmt := range n.Stmts {
				global, ok := stmt.(*ast.DeclarationStmt)
				if !ok {
					p.Error(n.Token(), "Invalid global variable: %s", n)
				}

				node.GlobalStmts = append(node.GlobalStmts, global)
			}
		case *ast.FuncDefNode:
			node.FuncDefs = append(node.FuncDefs, n)
		default:
			p.Error(n.Token(), "Unexpected top level token: '%s' of type '%s'", n)
		}

	}
	return node
}

func (p *Parser) global() ast.Node {
	start := p.cur
	baseTy := p.declspec()
	ty, identTkn := p.declarator(baseTy)
	if p.cur.Kind == token.LPAREN {
		return p.funcdef(ty, identTkn)
	}

	p.backTo(start)
	return p.declarationStmt(false)
}

func (p *Parser) funcdef(ty types.Type, identTkn *token.Token) *ast.FuncDefNode {
	p.curFn = ast.NewFuncDefNode(p.cur)
	p.curFn.Type = ty
	p.curFn.Name = identTkn.Str
	p.curFn.Args = p.funcdefargs()

	// Defined prior to parsing body in order to be called recursively.
	p.funcdefs[p.curFn.Name] = p.curFn
	p.curFn.Body = p.blockStmt()
	return p.curFn
}

func (p *Parser) funcdefargs() *ast.FuncDefArgs {
	p.expect(p.cur, token.LPAREN)
	lv := ast.NewLocalVariableNode(p.cur)
	args := &ast.FuncDefArgs{LV: lv}
	p.nextTkn()

	if p.cur.Kind == token.RPAREN {
		p.nextTkn()
		return args
	}

	basety1 := p.declspec()
	ty1, identTok := p.declarator(basety1)
	arg1 := &ast.LocalVariable{Name: identTok.Str, Type: ty1, IsLocal: true}
	args.LV.Locals = append(args.LV.Locals, arg1)

	for p.cur.Kind == token.COMMA {
		p.nextTkn()
		basety := p.declspec()
		ty, identTok := p.declarator(basety)
		arg := &ast.LocalVariable{Name: identTok.Str, Type: ty, IsLocal: true}
		args.LV.Locals = append(args.LV.Locals, arg)
	}

	p.expect(p.cur, token.RPAREN)
	p.nextTkn()

	p.prepareLocals(args.LV.Locals)
	return args
}

func (p *Parser) stmt() ast.Stmt {
	if p.cur.Kind == token.TYPE {
		return p.declarationStmt(true)
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
	node := &ast.ExpStmt{Exp: exp}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()
	return node
}

func (p *Parser) declspec() types.Type {
	debug("declspec")
	p.expect(p.cur, token.TYPE)

	tkn := p.cur
	p.nextTkn()

	switch tkn.Str {
	case "int":
		return types.GetInt()
	case "char":
		return types.GetChar()
	}

	p.Error(tkn, "Invalid type %s", tkn.Str)
	return nil
}

// declarator = "*"* ident ("[" num "]")?
func (p *Parser) declarator(ty types.Type) (types.Type, *token.Token) {
	debug("declarator")
	for p.cur.Kind == token.ASTERISK {
		ty = types.PointerTo(ty)
		p.nextTkn()
	}

	if p.cur.Kind != token.IDENT {
		return ty, nil
	}

	identTok := p.cur
	p.nextTkn()

	if p.cur.Kind == token.LBRACKET {
		p.nextTkn() // [
		p.expect(p.cur, token.NUM)
		length, err := strconv.Atoi(p.cur.Str)
		if err != nil || length <= 0 {
			p.tzer.Error(p.cur, "a positive number is expected. got %s.", p.cur.Str)
			os.Exit(1)
		}
		p.nextTkn()
		ty = types.ArrayOf(ty, length)
		p.expect(p.cur, token.RBRACKET)
		p.nextTkn()
	}

	return ty, identTok
}

func (p *Parser) declarationStmt(isLocal bool) *ast.StmtListNode {
	debug("declarationStmt")
	initTok := p.cur
	baseTy := p.declspec() // "int"

	locals := []*ast.LocalVariable{}
	stmts := []*ast.DeclarationStmt{}

	// int a,b,c = 0, d = 3;
	for p.cur.Kind != token.SEMICOLLON {
		if p.cur.Kind == token.COMMA {
			p.nextTkn()
		}

		ty, identTok := p.declarator(baseTy) // "**a"

		local := &ast.LocalVariable{Name: identTok.Str, Type: ty, IsLocal: isLocal}
		locals = append(locals, local)

		if p.cur.Kind != token.ASSIGN {
			continue
		}

		p.nextTkn() // "="
		p.prepareLocals(locals)

		left := ast.NewLocalVariableNode(initTok)
		left.Locals = locals
		right := p.expr()
		declStmt := ast.NewDeclarationStmt(left, right, "=", initTok)
		err := declStmt.CheckTypeError()
		if err != nil {
			p.Error(declStmt.Token(), err.Error())
		}
		stmts = append(stmts, declStmt)
		locals = []*ast.LocalVariable{}
	}

	if len(locals) > 0 {
		p.prepareLocals(locals)

		left := ast.NewLocalVariableNode(initTok)
		left.Locals = locals
		declStmt := ast.NewDeclarationStmt(left, nil, "=", initTok)
		err := declStmt.CheckTypeError()
		if err != nil {
			p.Error(declStmt.Token(), err.Error())
		}
		stmts = append(stmts, declStmt)
	}

	stmtList := &ast.StmtListNode{}
	stmtList.Stmts = []ast.Stmt{}
	for _, stmt := range stmts {
		stmtList.Stmts = append(stmtList.Stmts, stmt)
	}
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()
	return stmtList
}

func (p *Parser) blockStmt() *ast.BlockStmt {
	p.expect(p.cur, token.LBRACE)
	tkn := p.cur
	p.nextTkn() // {
	node := ast.NewBlockStmt(tkn)
	stmtList := &ast.StmtListNode{Stmts: []ast.Stmt{}}
	for p.cur.Kind != token.RBRACE {
		stmtList.Stmts = append(stmtList.Stmts, p.stmt())
	}
	node.Stmts = stmtList
	p.nextTkn() // }
	return node
}

func (p *Parser) forStmt() *ast.ForStmt {
	p.expect(p.cur, token.FOR)
	tkn := p.cur
	node := ast.NewForStmt(tkn)
	p.nextTkn()
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()

	if p.cur.Kind != token.SEMICOLLON {
		if p.cur.Kind == token.TYPE {
			node.Init = p.declarationStmt(true)
		} else {
			node.Init = p.expr()
			p.expect(p.cur, token.SEMICOLLON)
			p.nextTkn()
		}
	} else {
		p.expect(p.cur, token.SEMICOLLON)
		p.nextTkn()
	}

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

func (p *Parser) whileStmt() *ast.WhileStmt {
	p.expect(p.cur, token.WHILE)
	tkn := p.cur
	p.nextTkn()
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()
	exp := p.expr()
	p.expect(p.cur, token.RPAREN)
	p.nextTkn()
	body := p.stmt()
	return ast.NewWhileStmt(exp, body, tkn)
}

func (p *Parser) ifStmt() *ast.IfStmt {
	p.expect(p.cur, token.IF)
	tkn := p.cur
	p.nextTkn()
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()
	exp := p.expr()
	p.expect(p.cur, token.RPAREN)
	p.nextTkn()
	ifBody := p.stmt()
	node := ast.NewIfStmt(exp, ifBody, nil, tkn)
	if p.cur.Kind == token.ELSE {
		p.nextTkn()
		node.ElseBody = p.stmt()
	}
	return node
}

func (p *Parser) returnStmt() *ast.ReturnStmt {
	p.expect(p.cur, token.RETURN)
	tkn := p.cur
	p.nextTkn()
	exp := p.expr()
	node := ast.NewReturnStmt(exp, tkn)
	p.expect(p.cur, token.SEMICOLLON)
	p.nextTkn()
	return node
}

func (p *Parser) expr() ast.Exp {
	debug("expr")
	return p.assign()
}

func (p *Parser) assign() ast.Exp {
	debug("assign")
	node := p.eq()

	if p.cur.Kind == token.ASSIGN {
		infix := ast.NewInfixExp(node, nil, p.cur.Str, p.cur)
		p.nextTkn() // =
		infix.Right = p.assign()
		node = infix

		// TODO: duplicate left value check
		if ident, ok := infix.Left.(*ast.IdentExp); ok {
			_ = p.getDef(ident.Name)
		}

		err := infix.CheckTypeError()
		if err != nil {
			p.Error(node.Token(), err.Error())
		}
	}

	return node
}

func (p *Parser) eq() ast.Exp {
	debug("eq")
	node := p.lg()

	if p.cur.Kind == token.EQ || p.cur.Kind == token.NEQ {
		infix := ast.NewInfixExp(node, nil, p.cur.Str, p.cur)
		p.nextTkn()
		infix.Right = p.lg()
		node = infix
		err := infix.CheckTypeError()
		if err != nil {
			p.Error(node.Token(), err.Error())
		}
	}

	return node
}

func (p *Parser) lg() ast.Exp {
	debug("lg")
	node := p.add()

	switch p.cur.Kind {
	case token.LT:
		fallthrough
	case token.GT:
		fallthrough
	case token.LTE:
		fallthrough
	case token.GTE:
		infix := ast.NewInfixExp(node, nil, p.cur.Str, p.cur)
		p.nextTkn()
		infix.Right = p.add()
		node = infix
		err := infix.CheckTypeError()
		if err != nil {
			p.Error(node.Token(), err.Error())
		}
	}

	return node
}

func (p *Parser) add() ast.Exp {
	debug("add")
	node := p.mul()

	for p.cur.Kind == token.PLUS || p.cur.Kind == token.MINUS {
		switch p.cur.Kind {
		case token.PLUS:
			fallthrough
		case token.MINUS:
			infix := ast.NewInfixExp(node, nil, p.cur.Str, p.cur)
			p.nextTkn()
			infix.Right = p.mul()
			node = infix
			err := infix.CheckTypeError()
			if err != nil {
				p.Error(node.Token(), err.Error())
			}
		default:
			// never go here
			p.tzer.Error(p.cur, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) mul() ast.Exp {
	debug("mul")
	node := p.unary()

	for p.cur.Kind == token.ASTERISK || p.cur.Kind == token.SLASH {
		switch p.cur.Kind {
		case token.ASTERISK:
			fallthrough
		case token.SLASH:
			infix := ast.NewInfixExp(node, nil, p.cur.Str, p.cur)
			p.nextTkn()
			infix.Right = p.unary()
			node = infix
			err := infix.CheckTypeError()
			if err != nil {
				p.Error(node.Token(), err.Error())
			}
		default:
			// never go here
			p.tzer.Error(p.cur, "Invalid token: %s", p.cur.Str)
		}
	}

	return node
}

func (p *Parser) unary() ast.Exp {
	debug("unary")
	switch p.cur.Kind {
	case token.PLUS:
		fallthrough
	case token.MINUS:
		fallthrough
	case token.ASTERISK:
		fallthrough
	case token.SIZEOF:
		fallthrough
	case token.AND:
		node := ast.NewUnaryExp(nil, p.cur.Str, p.cur)
		p.nextTkn()
		node.Right = p.primary()
		return node
	default:
		n := p.primary()
		return n
	}
}

func (p *Parser) primary() ast.Exp {
	debug("primary")
	switch p.cur.Kind {
	case token.NUM:
		return p.num()
	case token.STRING:
		return p.str()
	case token.IDENT:
		exp := p.ident()
		ident, ok := exp.(*ast.IdentExp)
		if !ok {
			// funccall
			return exp
		}

		if p.cur.Kind == token.LBRACKET {
			p.nextTkn() // [
			index := p.expr()
			p.expect(p.cur, token.RBRACKET)
			p.nextTkn() // ]
			_ = p.getDef(ident.Name)
			return ast.NewIndexExp(ident, index, ident.Token())
		}

		return ident
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
		p.tzer.Error(p.cur, "Invalid token as primary: %s", p.cur.Str)
		return nil
	}
}

func (p *Parser) num() ast.Exp {
	p.expect(p.cur, token.NUM)
	node := ast.NewNumExp(p.cur.Val, p.cur)
	p.nextTkn()
	return node
}

func (p *Parser) str() ast.Exp {
	p.expect(p.cur, token.STRING)
	lbl := p.getLbl()
	node := ast.NewStringLiteralExp(p.cur.Str, p.cur, lbl)
	p.Strings = append(p.Strings, node)
	p.nextTkn()
	return node
}

func (p *Parser) ident() ast.Exp {
	p.expect(p.cur, token.IDENT)
	tkn := p.cur
	p.nextTkn()

	if p.cur.Kind == token.LPAREN {
		return p.funccall(tkn)
	} else {
		local := p.getDef(tkn.Str)
		return ast.NewIdentExp(tkn.Str, tkn, local.Type)
	}
}

func (p *Parser) funccall(identTkn *token.Token) *ast.FuncCallExp {
	p.expect(identTkn, token.IDENT)
	p.expect(p.cur, token.LPAREN)
	p.nextTkn()

	def, ok := p.funcdefs[identTkn.Str]
	if !ok {
		// コンパイル後にリンクされるので問題無し.
	}

	exp := ast.NewFuncCallExp(identTkn.Str, nil, identTkn, def)

	if p.cur.Kind == token.RPAREN {
		p.nextTkn()
		return exp
	}

	exp.Params = p.funccallparams()
	return exp
}

func (p *Parser) funccallparams() *ast.FuncCallParams {
	params := &ast.FuncCallParams{Exps: []ast.Exp{}}
	param1 := p.expr()
	params.Exps = append(params.Exps, param1)
	for p.cur.Kind == token.COMMA {
		p.nextTkn()
		param := p.expr()
		params.Exps = append(params.Exps, param)
	}

	p.expect(p.cur, token.RPAREN)
	p.nextTkn()
	return params
}

func (p *Parser) prepareLocals(locals []*ast.LocalVariable) {
	for _, local := range locals {
		if local.IsLocal {
			if _, exists := p.curFn.Offsets[local.Name]; exists {
				p.tzer.Error(p.cur, "Local variable already declared: %s", p.cur.Str)
			}

			p.curFn.OffsetCnt += local.Type.StackSize()
			p.curFn.Offsets[local.Name] = p.curFn.OffsetCnt
			p.curFn.Locals[local.Name] = local
		} else {
			if _, exists := p.Globals[local.Name]; exists {
				p.tzer.Error(p.cur, "Global variable already declared: %s", p.cur.Str)
			}

			p.Globals[local.Name] = local
		}
	}
}

func (p *Parser) getLbl() string {
	return fmt.Sprintf(".L.string.%d", p.strCnt)
}

func (p *Parser) expect(token *token.Token, kinds ...token.TokenKind) {
	p.tzer.Expect(token, kinds...)
}

func debug(s string, args ...interface{}) {
	if DEBUG {
		err(s, args...)
	}
}

func err(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

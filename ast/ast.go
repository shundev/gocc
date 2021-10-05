package ast

import (
	"bytes"
	"fmt"
	"go9cc/token"
	"go9cc/types"
	"os"
	"strings"
)

type LocalVariable struct {
	Name    string
	Type    types.Type
	IsLocal bool
	offset  int
}

func (n *LocalVariable) String() string {
	return n.Type.String() + " " + n.Name
}

type Node interface {
	String() string
	Token() *token.Token
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

type TypeError struct {
	msg string
}

func (t *TypeError) Error() string {
	return t.msg
}

/* Stmt List Stmt */

type StmtListNode struct {
	Stmts []Stmt
}

func (n *StmtListNode) stmtNode() {}

func (n *StmtListNode) Token() *token.Token {
	if len(n.Stmts) == 0 {
		return nil
	}

	return n.Stmts[0].Token()
}

func (n *StmtListNode) String() string {
	if len(n.Stmts) == 0 {
		return ""
	}

	var out bytes.Buffer
	ss := []string{}
	for _, stmt := range n.Stmts {
		ss = append(ss, stmt.String())
	}

	out.WriteString(strings.Join(ss, " "))
	return out.String()
}

/* Local Variable */

type LocalVariableNode struct {
	Locals []*LocalVariable
	token  *token.Token
}

func NewLocalVariableNode(token *token.Token) *LocalVariableNode {
	return &LocalVariableNode{
		Locals: []*LocalVariable{},
		token:  token,
	}
}

func (n *LocalVariableNode) Token() *token.Token {
	return n.token
}

func (n *LocalVariableNode) String() string {
	var out bytes.Buffer

	ss := []string{}
	for _, local := range n.Locals {
		s := local.String()
		ss = append(ss, s)
	}
	out.WriteString(strings.Join(ss, ", "))
	return out.String()
}

type FuncDefArgs struct {
	LV *LocalVariableNode
}

func (n *FuncDefArgs) String() string {
	var out bytes.Buffer

	ss := []string{}
	for _, local := range n.LV.Locals {
		ss = append(ss, local.String())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(ss, ", "))
	out.WriteString(")")
	return out.String()
}

/* Num */

type NumExp struct {
	Val   int
	token *token.Token
}

func NewNumExp(val int, token *token.Token) *NumExp {
	return &NumExp{
		Val: val, token: token,
	}
}

func (n *NumExp) expNode() {}

func (n *NumExp) Token() *token.Token {
	return n.token
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

func NewInfixExp(left, right Exp, op string, token *token.Token) *InfixExp {
	return &InfixExp{
		Left: left, Right: right, Op: op, token: token,
	}
}

func (n *InfixExp) expNode() {}

func (n *InfixExp) Token() *token.Token {
	return n.token
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

func (n *InfixExp) CheckTypeError() error {
	if n.Op == "=" {
		return nil
	}

	ret := &TypeError{}
	_, ok := n.Right.Type().(*types.Int)
	if !ok {
		ret.msg = fmt.Sprintf("Rvalue of arithmetic must be int, but got %T in exp %s", n.Right.Type(), n)
		return ret
	}

	if n.Op == "*" || n.Op == "/" {
		_, ok := n.Left.Type().(*types.Int)
		if !ok {
			ret.msg = fmt.Sprintf("Pointer cannot be multiplied nor divided: %s", n)
			return ret
		}
	}

	return nil
}

/* Declaration */

type DeclarationStmt struct {
	LV    *LocalVariableNode
	Exp   Exp
	Op    string
	token *token.Token
}

func NewDeclarationStmt(lv *LocalVariableNode, exp Exp, op string, token *token.Token) *DeclarationStmt {
	return &DeclarationStmt{LV: lv, Exp: exp, Op: op, token: token}
}

func (n *DeclarationStmt) stmtNode() {}

func (n *DeclarationStmt) Token() *token.Token {
	return n.token
}

func (n *DeclarationStmt) String() string {
	var out bytes.Buffer
	out.WriteString(n.LV.String())
	if n.Exp != nil {
		out.WriteString(" " + n.Op + " ")
		out.WriteString(n.Exp.String())
	}
	out.WriteString(";")
	return out.String()
}

func (n *DeclarationStmt) CheckTypeError() error {
	if n.Exp == nil {
		return nil
	}

	ret := &TypeError{}
	for _, local := range n.LV.Locals {
		if n.Exp.Type().String() != local.Type.String() {
			ret.msg = fmt.Sprintf("Type mismatch: %s", local)
			return ret
		}
	}

	return nil
}

/* Unary */

type UnaryExp struct {
	Right Exp
	Op    string
	token *token.Token
}

func NewUnaryExp(right Exp, op string, token *token.Token) *UnaryExp {
	return &UnaryExp{
		Right: right,
		Op:    op,
		token: token,
	}
}

func (n *UnaryExp) expNode() {}

func (n *UnaryExp) Token() *token.Token {
	return n.token
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
		return types.PointerTo(n.Right.Type())
	case "sizeof":
		return types.GetInt()
	}

	err("Invalid op: %s", n.Op)
	os.Exit(1)
	return types.GetInt()
}

/* Identifier */

type IdentExp struct {
	Name  string
	token *token.Token
	typ   types.Type
}

func NewIdentExp(name string, token *token.Token, typ types.Type) *IdentExp {
	return &IdentExp{
		Name: name, token: token, typ: typ,
	}
}

func (n *IdentExp) expNode() {}

func (n *IdentExp) Token() *token.Token {
	return n.token
}

func (n *IdentExp) String() string {
	return n.Name
}

func (n *IdentExp) Type() types.Type {
	return n.typ
}

/* Array Index Access */

type IndexExp struct {
	Ident *IdentExp
	Index Exp
	token *token.Token
}

func NewIndexExp(ident *IdentExp, index Exp, token *token.Token) *IndexExp {
	return &IndexExp{
		Ident: ident,
		Index: index,
		token: token,
	}
}

func (n *IndexExp) expNode() {}

func (n *IndexExp) Token() *token.Token {
	return n.token
}

func (n *IndexExp) String() string {
	var out bytes.Buffer
	out.WriteString(n.Ident.Name)
	out.WriteString("[")
	out.WriteString(n.Index.String())
	out.WriteString("]")
	return out.String()
}

func (n *IndexExp) Type() types.Type {
	arrayTyp, ok := n.Ident.Type().(*types.Array)
	if !ok {
		err("Array type expected, but got=%s", n.Ident.Type())
	}
	return arrayTyp.Base
}

/* Func Call Params */

type FuncCallParams struct {
	Exps []Exp
}

func NewFuncCallParams(exps []Exp) *FuncCallParams {
	if exps == nil {
		exps = []Exp{}
	}
	return &FuncCallParams{Exps: exps}
}

func (n *FuncCallParams) Token() *token.Token {
	if len(n.Exps) == 0 {
		return nil
	}

	return n.Exps[0].Token()
}

func (n *FuncCallParams) String() string {
	var out bytes.Buffer
	ss := []string{}
	for _, exp := range n.Exps {
		ss = append(ss, exp.String())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(ss, ", "))
	out.WriteString(")")
	return out.String()
}

/* Function */

type FuncCallExp struct {
	Name   string
	Params *FuncCallParams
	token  *token.Token
	Def    *FuncDefNode
}

func NewFuncCallExp(name string, params *FuncCallParams, token *token.Token, def *FuncDefNode) *FuncCallExp {
	if params == nil {
		params = NewFuncCallParams(nil)
	}

	node := &FuncCallExp{
		Name:   name,
		Params: params,
		Def:    def,
		token:  token,
	}

	return node
}

func (n *FuncCallExp) expNode() {}

func (n *FuncCallExp) Token() *token.Token {
	return n.token
}

func (n *FuncCallExp) Type() types.Type {
	return n.Def.Type
}

func (n *FuncCallExp) String() string {
	var out bytes.Buffer
	out.WriteString(n.Name)
	out.WriteString(n.Params.String())
	return out.String()
}

/* Statement */

type ExpStmt struct {
	Exp   Exp
	token *token.Token
}

func (n *ExpStmt) stmtNode() {}

func (n *ExpStmt) Token() *token.Token {
	return n.token
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

func NewReturnStmt(exp Exp, token *token.Token) *ReturnStmt {
	return &ReturnStmt{
		Exp:   exp,
		token: token,
	}
}

func (n *ReturnStmt) stmtNode() {}

func (n *ReturnStmt) Token() *token.Token {
	return n.token
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

func NewIfStmt(cond Exp, ifBody Stmt, elseBody Stmt, token *token.Token) *IfStmt {
	return &IfStmt{
		Cond:     cond,
		IfBody:   ifBody,
		ElseBody: elseBody,
		token:    token,
	}
}

func (n *IfStmt) stmtNode() {}

func (n *IfStmt) Token() *token.Token {
	return n.token
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

func NewWhileStmt(cond Exp, body Stmt, token *token.Token) *WhileStmt {
	return &WhileStmt{
		Cond:  cond,
		Body:  body,
		token: token,
	}
}

func (n *WhileStmt) stmtNode() {}

func (n *WhileStmt) Token() *token.Token {
	return n.token
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
	Init            Node
	Cond, AfterEach Exp
	Body            Stmt
	token           *token.Token
}

func NewForStmt(token *token.Token) *ForStmt {
	return &ForStmt{token: token}
}

func (n *ForStmt) stmtNode() {}

func (n *ForStmt) Token() *token.Token {
	return n.token
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
	Stmts *StmtListNode
	token *token.Token
}

func NewBlockStmt(token *token.Token) *BlockStmt {
	return &BlockStmt{token: token}
}

func (n *BlockStmt) stmtNode() {}

func (n *BlockStmt) Token() *token.Token {
	return n.token
}

func (n *BlockStmt) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	out.WriteString(n.Stmts.String())
	out.WriteString(" }")
	return out.String()
}

/* Program */

type ProgramNode struct {
	FuncDefs    []*FuncDefNode
	GlobalStmts []*DeclarationStmt
}

func (n *ProgramNode) Token() *token.Token {
	if len(n.FuncDefs) > 0 {
		return n.FuncDefs[0].Token()
	}

	return nil
}

func (n *ProgramNode) String() string {
	ss := []string{}
	for _, stmt := range n.GlobalStmts {
		ss = append(ss, stmt.String())
	}
	for _, stmt := range n.FuncDefs {
		ss = append(ss, stmt.String())
	}
	return strings.Join(ss, " ")
}

/* Func Def */

type FuncDefNode struct {
	Body      *BlockStmt
	Name      string
	Type      types.Type
	Offsets   map[string]int
	StackSize int
	Args      *FuncDefArgs
	OffsetCnt int
	Locals    map[string]*LocalVariable
	token     *token.Token
}

func NewFuncDefNode(token *token.Token) *FuncDefNode {
	return &FuncDefNode{
		token:   token,
		Offsets: map[string]int{},
		Locals:  map[string]*LocalVariable{},
	}
}

func (n *FuncDefNode) Token() *token.Token {
	return n.token
}

func (n *FuncDefNode) String() string {
	var out bytes.Buffer
	out.WriteString(n.Type.String())
	out.WriteString(" ")
	out.WriteString(n.Name)
	out.WriteString(" ")
	out.WriteString(n.Args.String())
	out.WriteString(" ")
	out.WriteString(n.Body.String())
	return out.String()
}

func (n *FuncDefNode) PrepareStackSize() {
	max := 0
	for _, v := range n.Offsets {
		if v > max {
			max = v
		}
	}

	n.StackSize = alignTo(max, 16)
}

func err(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n")
}

func alignTo(n, align int) int {
	return (n + align - 1) / align * align
}

package generator

import (
	"fmt"
	"go9cc/parser"
	"io"
	"os"
)

const HEADER = `.intel_syntax noprefix
.globl main
main:
`

const (
	RAX = "rax"
	RDI = "rdi"
	RBP = "rbp" // base pointer
	RSP = "rsp" // stack pointer
	AL  = "al"
)

type Generator struct {
	parser *parser.Parser
	out    io.Writer
	lblCnt int
}

func New(p *parser.Parser, out io.Writer) *Generator {
	gen := &Generator{parser: p, out: out}
	return gen
}

func (g *Generator) Gen() {
	g.header()

	node := g.parser.Parse()
	g.prolog()
	g.walk(node)
	g.epilog()
}

// push corresponding address to the top of stack
func (g *Generator) lvalue(node parser.Node) {
	ident, ok := node.(*parser.IdentExp)
	if !ok {
		fmt.Fprintf(os.Stderr, "Lvalue must be a ident node, but got: %T, %s\n", node, node.String())
		os.Exit(1)
	}

	offset := getOffset(ident)

	g.mov(RAX, RBP)
	g.sub(RAX, fmt.Sprintf("%d", offset))
	g.push(RAX)
}

/*
nodeの評価結果をスタックトップにpushする
*/
func (g *Generator) walk(node parser.Node) {
	switch node.(type) {
	case *parser.ProgramNode:
		program, _ := node.(*parser.ProgramNode)
		for _, stmt := range program.Stmts {
			g.walk(stmt)
		}
	case *parser.ExpStmt:
		stmt, _ := node.(*parser.ExpStmt)
		g.walk(stmt.Exp)
		g.pop(RAX)
	case *parser.ReturnStmt:
		stmt, _ := node.(*parser.ReturnStmt)
		g.walk(stmt.Exp)
		g.pop(RAX)
		g.epilog()
	case *parser.BlockStmt:
		block, _ := node.(*parser.BlockStmt)
		for _, stmt := range block.Stmts {
			g.walk(stmt)
		}
	case *parser.ForStmt:
		stmt, _ := node.(*parser.ForStmt)
		lblBegin := g.genLbl()
		lblEnd := g.genLbl()
		if stmt.Init != nil {
			g.walk(stmt.Init)
		}
		g.label(lblBegin)
		if stmt.Cond != nil {
			g.walk(stmt.Cond)
			g.pop(RAX)
			g.cmp(RAX, "0")
			g.je(lblEnd) // RAXが0(false)ならforの外にジャンプ
		}
		g.walk(stmt.Body)
		if stmt.AfterEach != nil {
			g.walk(stmt.AfterEach)
		}
		g.jmp(lblBegin)
		g.label(lblEnd)
	case *parser.WhileStmt:
		stmt, _ := node.(*parser.WhileStmt)
		lblBegin := g.genLbl()
		lblEnd := g.genLbl()
		g.label(lblBegin)
		g.walk(stmt.Cond)
		g.pop(RAX)
		g.cmp(RAX, "0")
		g.je(lblEnd) // RAXが0(false)ならwhileの外にジャンプ
		g.walk(stmt.Body)
		g.jmp(lblBegin)
		g.label(lblEnd)
	case *parser.IfStmt:
		stmt, _ := node.(*parser.IfStmt)
		lblElse := g.genLbl()
		lblEnd := g.genLbl()
		g.walk(stmt.Cond)
		g.pop(RAX)
		g.cmp(RAX, "0")
		g.je(lblElse) // RAXが0(false)ならelseブロックにジャンプ
		g.walk(stmt.IfBody)
		g.jmp(lblEnd)
		g.label(lblElse)
		if stmt.ElseBody != nil {
			g.walk(stmt.ElseBody)
		}
		g.label(lblEnd)
	case *parser.NumExp:
		num, _ := node.(*parser.NumExp)
		g.push(fmt.Sprintf("%d", num.Val))
	case *parser.IdentExp:
		// 変数呼び出し
		ident, _ := node.(*parser.IdentExp)
		offset := getOffset(ident)
		g.mov(RAX, RBP)
		g.sub(RAX, fmt.Sprintf("%d", offset))
		g.mov(RAX, "["+RAX+"]")
		g.push(RAX)
	case *parser.FuncCallExp:
		// 変数呼び出し
		node, _ := node.(*parser.FuncCallExp)
		g.call(node.Name)
	case *parser.UnaryExp:
		// Regard unary as infix for easy development(i.e. -1 -> 0 - 1)
		unary, _ := node.(*parser.UnaryExp)
		infix := parser.UnaryToInfix(unary)
		g.walk(infix)
	case *parser.InfixExp:
		infix, _ := node.(*parser.InfixExp)
		if infix.Op == "=" {
			g.lvalue(infix.Left)
			g.walk(infix.Right)
			g.pop(RDI)
			g.pop(RAX)
			g.mov("["+RAX+"]", RDI) // Write RDI to memory RAX
			g.push(RDI)             // eval a = 2 to 2
			return
		}

		g.walk(infix.Left)
		g.walk(infix.Right)

		// <と>は左右だけ入れ替えて同じアセンブリ命令setlを使う
		if infix.Op == ">" || infix.Op == ">=" {
			g.pop(RAX)
			g.pop(RDI)
		} else {
			g.pop(RDI)
			g.pop(RAX)
		}

		switch infix.Op {
		case "+":
			g.add(RAX, RDI)
		case "-":
			g.sub(RAX, RDI)
		case "*":
			g.mul(RAX, RDI)
		case "/":
			g.div(RDI)
		case "<":
			fallthrough
		case ">":
			g.cmp(RAX, RDI)
			g.setl(AL)
			g.movzb(RAX, AL)
		case "<=":
			fallthrough
		case ">=":
			g.cmp(RAX, RDI)
			g.setle(AL)
			g.movzb(RAX, AL)
		case "==":
			g.cmp(RAX, RDI)
			g.sete(AL)
			g.movzb(RAX, AL)
		case "!=":
			g.cmp(RAX, RDI)
			g.setne(AL)
			g.movzb(RAX, AL)
		}
		g.push(RAX)
	default:
		fmt.Fprintf(os.Stderr, "Unknown node: %T, %s\n", node, node.String())
		os.Exit(1)
	}
}

func (g *Generator) header() {
	io.WriteString(g.out, HEADER)
}

func (g *Generator) prolog() {
	numArgs := 26
	argSize := 8
	g.push(RBP)
	g.mov(RBP, RSP)
	g.sub(RSP, fmt.Sprintf("%d", numArgs*argSize))
}

func (g *Generator) epilog() {
	g.mov(RSP, RBP)
	g.pop(RBP)
	g.ret()
}

func (g *Generator) mov(rad string, val string) {
	s := fmt.Sprintf("  mov %s, %s\n", rad, val)
	io.WriteString(g.out, s)
}

func (g *Generator) add(rad1, rad2 string) {
	s := fmt.Sprintf("  add %s, %s\n", rad1, rad2)
	io.WriteString(g.out, s)
}

func (g *Generator) sub(rad1, rad2 string) {
	s := fmt.Sprintf("  sub %s, %s\n", rad1, rad2)
	io.WriteString(g.out, s)
}

func (g *Generator) mul(rad1, rad2 string) {
	s := fmt.Sprintf("  imul %s, %s\n", rad1, rad2)
	io.WriteString(g.out, s)
}

func (g *Generator) div(rad string) {
	io.WriteString(g.out, "  cqo\n")     // RAXのコードを伸ばしてRDX/RAXにセットする
	s := fmt.Sprintf("  idiv %s\n", rad) // RDX/RAXを128bitとみなして`rad`のレジスタの値で符号付除算
	io.WriteString(g.out, s)
}

func (g *Generator) push(val string) {
	s := fmt.Sprintf("  push %s\n", val)
	io.WriteString(g.out, s)
}

func (g *Generator) pop(rad string) {
	s := fmt.Sprintf("  pop %s\n", rad)
	io.WriteString(g.out, s)
}

func (g *Generator) sete(rad1 string) {
	s := fmt.Sprintf("  sete %s\n", rad1)
	io.WriteString(g.out, s)
}

func (g *Generator) setne(rad1 string) {
	s := fmt.Sprintf("  setne %s\n", rad1)
	io.WriteString(g.out, s)
}

func (g *Generator) setl(rad1 string) {
	s := fmt.Sprintf("  setl %s\n", rad1)
	io.WriteString(g.out, s)
}

func (g *Generator) setle(rad1 string) {
	s := fmt.Sprintf("  setle %s\n", rad1)
	io.WriteString(g.out, s)
}

func (g *Generator) je(label string) {
	s := fmt.Sprintf("  je .%s\n", label)
	io.WriteString(g.out, s)
}

func (g *Generator) jne(label string) {
	s := fmt.Sprintf("  jne .%s\n", label)
	io.WriteString(g.out, s)
}

func (g *Generator) jmp(label string) {

	s := fmt.Sprintf("  jmp .%s\n", label)
	io.WriteString(g.out, s)
}

func (g *Generator) call(label string) {

	s := fmt.Sprintf("  call %s\n", label)
	io.WriteString(g.out, s)
}

func (g *Generator) cmp(rad1, rad2 string) {
	s := fmt.Sprintf("  cmp %s, %s\n", rad1, rad2)
	io.WriteString(g.out, s)
}

func (g *Generator) movzb(rad1, rad2 string) {
	s := fmt.Sprintf("  movzb %s, %s\n", rad1, rad2)
	io.WriteString(g.out, s)
}

func (g *Generator) ret() {
	io.WriteString(g.out, "  ret\n")
}

func (g *Generator) label(name string) {
	s := fmt.Sprintf(".%s:\n", name)
	io.WriteString(g.out, s)
}

func (g *Generator) genLbl() string {
	s := fmt.Sprintf("L%d", g.lblCnt)
	g.lblCnt++
	return s
}

func getOffset(ident *parser.IdentExp) int {
	argSize := 8
	offset := -1
	mapping := "abcdefghijklmnopqrstuvwxyz"
	for i, r := range mapping {
		if string(r) == ident.Name {
			offset = i
		}
	}

	if offset == -1 {
		fmt.Fprintf(os.Stderr, "Invalid ident.Name: %s\n", ident.Name)
		os.Exit(1)
	}
	offset++
	return offset * argSize
}

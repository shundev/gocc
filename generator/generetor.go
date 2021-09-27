package generator

import (
	"fmt"
	"go9cc/parser"
	"go9cc/types"
	"io"
	"os"
)

const HEADER = `.intel_syntax noprefix
.globl main
`

const (
	RAX = "rax"
	RDI = "rdi"
	RBP = "rbp" // base pointer
	RSP = "rsp" // stack pointer
	AL  = "al"
)

type Generator struct {
	parser    *parser.Parser
	out       io.Writer
	lblCnt    int
	offsets   map[string]int
	stackSize int
	currentFn *parser.FuncDefNode
}

func New(p *parser.Parser, out io.Writer) *Generator {
	gen := &Generator{parser: p, out: out}
	return gen
}

func (g *Generator) Gen() {
	g.header()

	node := g.parser.Parse()
	g.offsets = node.Offsets
	g.stackSize = node.StackSize()

	g.walk(node)
}

// push corresponding address to the top of stack
func (g *Generator) address(node interface{}) {
	switch ty := node.(type) {
	case *parser.UnaryExp:
		// Nested unary.
		if ty.Op == "*" {
			g.walk(ty.Right)
			return
		}
	case *parser.LocalVariable:
		offset := g.getOffset(ty.Name)
		g.lea(RAX, RBP, -offset)
		return
	case *parser.IdentExp:
		offset := g.getOffset(ty.Name)
		g.lea(RAX, RBP, -offset)
		return
	}

	fmt.Fprintf(os.Stderr, "address must be a ident node, but got: %T\n", node)
	os.Exit(1)
}

func (g *Generator) walk(node parser.Node) {
	switch ty := node.(type) {
	case *parser.ProgramNode:
		for _, stmt := range ty.FuncDefs {
			g.walk(stmt)
		}
	case *parser.ExpStmt:
		g.walk(ty.Exp)
	case *parser.ReturnStmt:
		g.walk(ty.Exp)
		s := fmt.Sprintf(".L.return.%s", g.currentFn.Name)
		g.jmp(s)
	case *parser.StmtListNode:
		for _, stmt := range ty.Stmts {
			g.walk(stmt)
		}
	case *parser.BlockStmt:
		g.walk(ty.Stmts)
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
		val := fmt.Sprintf("%d", ty.Val)
		g.mov(RAX, val)
	case *parser.IdentExp:
		// 変数呼び出し
		g.address(ty)
		g.mov(RAX, "["+RAX+"]")
	case *parser.FuncCallExp:
		// 関数呼び出し
		g.call(ty.Name)
	case *parser.FuncDefNode:
		g.currentFn = ty
		g.label(ty.Name)
		g.prolog(g.stackSize)
		g.walk(ty.Body)
		g.label(fmt.Sprintf(".L.return.%s", ty.Name))
		g.epilog()
	case *parser.UnaryExp:
		unary := ty
		switch unary.Op {
		case "&":
			g.address(unary.Right) // RAXに目標のアドレスが載る
		case "*":
			g.walk(unary.Right) // RAXに目標のアドレスが載る
			g.mov(RAX, "["+RAX+"]")
		case "+":
			// do nothing ( +5 -> 5)
		case "-":
			g.walk(unary.Right)
			g.neg(RAX)
		}
	case *parser.DeclarationExp:
		for _, local := range ty.LV.Locals {
			if ty.Exp != nil {
				g.address(local)
				g.push(RAX) // 直近2つのRAXが必要な場合は前のRAXをスタックに退避
				g.walk(ty.Exp)
				g.pop(RDI)
				g.mov("["+RDI+"]", RAX)
			}
		}
	case *parser.InfixExp:
		infix := ty
		if infix.Op == "=" {
			g.address(infix.Left)
			g.push(RAX) // 直近2つのRAXが必要な場合は前のRAXをスタックに退避
			g.walk(infix.Right)
			g.pop(RDI)
			g.mov("["+RDI+"]", RAX)
			return
		}

		// ポインタ演算はタイプのサイズによってスケールする
		if infix.Op == "+" || infix.Op == "-" {
			switch infix.Left.Type().(type) {
			case *types.IntPointer:
				infix = parser.Scale(infix)
			}
		}

		g.walk(infix.Right) // 先に計算した方がRDIに入るから右辺を先にしないと-の時問題
		g.push(RAX)
		g.walk(infix.Left)
		g.pop(RDI)

		switch infix.Op {
		case "+":
			g.add(RAX, RDI)
		case "-":
			g.sub(RAX, RDI) // 右辺をRDIに入れているから
		case "*":
			g.mul(RAX, RDI)
		case "/":
			g.div(RDI)
		case ">":
			// swap RAX and RDI
			g.push(RAX)
			g.mov(RAX, RDI)
			g.pop(RDI)
			fallthrough
		case "<":
			g.cmp(RAX, RDI)
			g.setl(AL)
			g.movzb(RAX, AL)
		case ">=":
			g.push(RAX)
			g.mov(RAX, RDI)
			g.pop(RDI)
			fallthrough
		case "<=":
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown node: %T, %s\n", node, node.String())
		os.Exit(1)
	}
}

func (g *Generator) header() {
	io.WriteString(g.out, HEADER)
}

func (g *Generator) prolog(stackSize int) {
	g.push(RBP)
	g.mov(RBP, RSP)
	g.sub(RSP, fmt.Sprintf("%d", stackSize))
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
	s := fmt.Sprintf("  je %s\n", label)
	io.WriteString(g.out, s)
}

func (g *Generator) jne(label string) {
	s := fmt.Sprintf("  jne %s\n", label)
	io.WriteString(g.out, s)
}

func (g *Generator) jmp(label string) {
	s := fmt.Sprintf("  jmp %s\n", label)
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

func (g *Generator) lea(rad1, rad2 string, offset int) {
	s := fmt.Sprintf("  lea %s, [%s%d]\n", rad1, rad2, offset)
	io.WriteString(g.out, s)
}

func (g *Generator) neg(rad1 string) {
	s := fmt.Sprintf("  neg %s\n", rad1)
	io.WriteString(g.out, s)
}

func (g *Generator) ret() {
	io.WriteString(g.out, "  ret\n")
}

func (g *Generator) label(name string) {
	s := fmt.Sprintf("%s:\n", name)
	io.WriteString(g.out, s)
}

func (g *Generator) genLbl() string {
	s := fmt.Sprintf(".L%d", g.lblCnt)
	g.lblCnt++
	return s
}

func (g *Generator) getOffset(name string) int {
	offset, ok := g.offsets[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid ident name: %s\n", name)
		os.Exit(1)
	}

	return g.stackSize - offset + 8
}

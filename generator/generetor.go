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
	AL  = "al"
)

type Generator struct {
	parser *parser.Parser
	out    io.Writer
}

func New(p *parser.Parser, out io.Writer) *Generator {
	gen := &Generator{parser: p, out: out}
	return gen
}

func (g *Generator) Gen() {
	g.header()
	defer g.ret()

	node := g.parser.Parse()
	g.walk(node)
	g.pop(RAX)
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
	case *parser.StmtNode:
		stmt, _ := node.(*parser.StmtNode)
		g.walk(stmt.Exp)
	case *parser.NumNode:
		num, _ := node.(*parser.NumNode)
		g.push(fmt.Sprintf("%d", num.Val))
	case *parser.UnaryNode:
		// Regard unary as infix for easy development(i.e. -1 -> 0 - 1)
		unary, _ := node.(*parser.UnaryNode)
		infix := parser.UnaryToInfix(unary)
		g.walk(infix)
	case *parser.InfixNode:
		infix, _ := node.(*parser.InfixNode)
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
		fmt.Fprintf(os.Stderr, "Unknown node: %T\n", node)
		os.Exit(1)
	}
}

func (g *Generator) header() {
	io.WriteString(g.out, HEADER)
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

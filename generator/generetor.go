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
	g.pop("rax")
}

/*
nodeの評価結果をスタックトップにpushする
*/
func (g *Generator) walk(node parser.Node) {
	switch node.(type) {
	case *parser.NumNode:
		num, _ := node.(*parser.NumNode)
		g.push(fmt.Sprintf("%d", num.Val))
	case *parser.InfixNode:
		infix, _ := node.(*parser.InfixNode)
		g.walk(infix.Left)
		g.walk(infix.Right)

		g.pop("rdi")
		g.pop("rax")
		switch infix.Op {
		case "+":
			g.add("rax", "rdi")
		case "-":
			g.sub("rax", "rdi")
		case "*":
			g.mul("rax", "rdi")
		case "/":
			g.div("rdi")
		}
		g.push("rax")
	default:
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

func (g *Generator) ret() {
	io.WriteString(g.out, "  ret\n")
}

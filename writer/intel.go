package writer

import (
	"bytes"
	"fmt"
	"io"
)

type Writer interface {
	Header()
	Commit()
	Mov(string, string)
	Add(string, string)
	Sub(string, string)
	Mul(string, string)
	Div(string)
	Lea(int, string, string)
	Push(string)
	Pop(string)
	Sete(string)
	Setne(string)
	Setl(rad1 string)
	Setle(rad1 string)
	Je(label string)
	Jne(label string)
	Jmp(label string)
	Call(label string)
	Cmp(rad1, rad2 string)
	Movzb(rad1, rad2 string)
	Neg(rad1 string)
	Ret()
	Label(name string)
	Address(name string) string
}

type Intel struct {
	out io.Writer
	buf *bytes.Buffer
}

const INTEL_HEADER = `.intel_syntax noprefix
.globl main
`

func NewIntelWriter(out io.Writer) *Intel {
	return &Intel{out: out, buf: &bytes.Buffer{}}
}

func (g *Intel) Commit() {
	io.WriteString(g.out, g.buf.String())
}

func (g *Intel) Header() {
	io.WriteString(g.buf, INTEL_HEADER)
}

func (g *Intel) Mov(rad1 string, rad2 string) {
	s := fmt.Sprintf("  mov %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Add(rad1, rad2 string) {
	s := fmt.Sprintf("  add %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Sub(rad1, rad2 string) {
	s := fmt.Sprintf("  sub %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Mul(rad1, rad2 string) {
	s := fmt.Sprintf("  imul %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Div(rad string) {
	io.WriteString(g.buf, "  cqo\n")     // RAXのコードを伸ばしてRDX/RAXにセットする
	s := fmt.Sprintf("  idiv %s\n", rad) // RDX/RAXを128bitとみなして`rad`のレジスタの値で符号付除算
	io.WriteString(g.buf, s)
}

func (g *Intel) Push(val string) {
	s := fmt.Sprintf("  push %s\n", val)
	io.WriteString(g.buf, s)
}

func (g *Intel) Pop(rad string) {
	s := fmt.Sprintf("  pop %s\n", rad)
	io.WriteString(g.buf, s)
}

func (g *Intel) Sete(rad1 string) {
	s := fmt.Sprintf("  sete %s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Setne(rad1 string) {
	s := fmt.Sprintf("  setne %s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Setl(rad1 string) {
	s := fmt.Sprintf("  setl %s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Setle(rad1 string) {
	s := fmt.Sprintf("  setle %s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Je(label string) {
	s := fmt.Sprintf("  je %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *Intel) Jne(label string) {
	s := fmt.Sprintf("  jne %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *Intel) Jmp(label string) {
	s := fmt.Sprintf("  jmp %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *Intel) Call(label string) {
	s := fmt.Sprintf("  call %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *Intel) Cmp(rad1, rad2 string) {
	s := fmt.Sprintf("  cmp %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Movzb(rad1, rad2 string) {
	s := fmt.Sprintf("  movzb %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Lea(offset int, rad1, rad2 string) {
	s := fmt.Sprintf("  lea %s, %d[%s]\n", rad2, offset, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Neg(rad1 string) {
	s := fmt.Sprintf("  neg %s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Ret() {
	io.WriteString(g.buf, "  ret\n")
}

func (g *Intel) Label(name string) {
	s := fmt.Sprintf("%s:\n", name)
	io.WriteString(g.buf, s)
}

func (g *Intel) Address(name string) string {
	return fmt.Sprintf("[%s]", name)
}

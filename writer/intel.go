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
	Lea(offset, rad1, rad2 string)
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
	Movsx(rad1, rad2 string)
	Neg(rad1 string)
	Ret()
	Globl(label string)
	Size(size int)
	Data()
	Label(name string)
	String(value string)
	Text(text string)
	Address(name string) string
	Index(base, unit string, size int) string
}

type Intel struct {
	out io.Writer
	buf *bytes.Buffer
}

const INTEL_HEADER = `.intel_syntax noprefix
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

func (g *Intel) Movsx(rad1, rad2 string) {
	s := fmt.Sprintf("  movsx %s, %s\n", rad2, rad1)
	io.WriteString(g.buf, s)
}

func (g *Intel) Lea(offset, rad1, rad2 string) {
	s := fmt.Sprintf("  lea %s, %s[%s]\n", rad2, offset, rad1)
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

func (g *Intel) Globl(label string) {
	g.Text(fmt.Sprintf(".globl %s", label))
}

func (g *Intel) String(value string) {
	g.Text(fmt.Sprintf(".string \"%s\"", value))
}

func (g *Intel) Size(value int) {
	g.Text(fmt.Sprintf(".size, %d", value))
}

func (g *Intel) Data() {
	g.Text(".data")
}

func (g *Intel) Text(text string) {
	s := fmt.Sprintf("  %s\n", text)
	io.WriteString(g.buf, s)
}

func (g *Intel) Address(name string) string {
	return fmt.Sprintf("[%s]", name)
}

func (g *Intel) Index(base, unit string, size int) string {
	return fmt.Sprintf("%s+%s*%d", base, unit, size)
}

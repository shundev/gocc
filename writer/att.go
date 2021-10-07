package writer

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type ATT struct {
	out io.Writer
	buf *bytes.Buffer
}

const header = ""

func New(out io.Writer, intel bool) Writer {
	if intel {
		return NewIntelWriter(out)
	}

	return NewATTWriter(out)
}

func NewATTWriter(out io.Writer) *ATT {
	return &ATT{out: out, buf: &bytes.Buffer{}}
}

func (g *ATT) Commit() {
	io.WriteString(g.out, g.buf.String())
}

func (g *ATT) Header() {
	io.WriteString(g.buf, header)
}

func (g *ATT) Mov(rad1 string, rad2 string) {
	s := fmt.Sprintf("  mov %s, %s\n", prefixed(rad1), prefixed(rad2))
	io.WriteString(g.buf, s)
}

func (g *ATT) Add(rad1, rad2 string) {
	s := fmt.Sprintf("  add %s, %%%s\n", prefixed(rad1), rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Sub(rad1, rad2 string) {
	s := fmt.Sprintf("  sub %s, %%%s\n", prefixed(rad1), rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Mul(rad1, rad2 string) {
	s := fmt.Sprintf("  imul %s, %%%s\n", prefixed(rad1), rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Div(rad string) {
	io.WriteString(g.buf, "  cqo\n")       // RAXのコードを伸ばしてRDX/RAXにセットする
	s := fmt.Sprintf("  idiv %%%s\n", rad) // RDX/RAXを128bitとみなして`rad`のレジスタの値で符号付除算
	io.WriteString(g.buf, s)
}

func (g *ATT) Push(val string) {
	s := fmt.Sprintf("  push %%%s\n", val)
	io.WriteString(g.buf, s)
}

func (g *ATT) Pop(rad string) {
	s := fmt.Sprintf("  pop %%%s\n", rad)
	io.WriteString(g.buf, s)
}

func (g *ATT) Sete(rad1 string) {
	s := fmt.Sprintf("  sete %%%s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *ATT) Setne(rad1 string) {
	s := fmt.Sprintf("  setne %%%s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *ATT) Setl(rad1 string) {
	s := fmt.Sprintf("  setl %%%s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *ATT) Setle(rad1 string) {
	s := fmt.Sprintf("  setle %%%s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *ATT) Je(label string) {
	s := fmt.Sprintf("  je %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *ATT) Jne(label string) {
	s := fmt.Sprintf("  jne %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *ATT) Jmp(label string) {
	s := fmt.Sprintf("  jmp %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *ATT) Call(label string) {
	s := fmt.Sprintf("  call %s\n", label)
	io.WriteString(g.buf, s)
}

func (g *ATT) Cmp(rad1, rad2 string) {
	s := fmt.Sprintf("  cmp %s, %%%s\n", prefixed(rad1), rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Movzb(rad1, rad2 string) {
	s := fmt.Sprintf("  movzb %s, %%%s\n", prefixed(rad1), rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Movsx(rad1, rad2 string) {
	s := fmt.Sprintf("  movsx %s, %%%s\n", prefixed(rad1), rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Lea(offset, rad1, rad2 string) {
	s := fmt.Sprintf("  lea %s(%%%s), %%%s\n", offset, rad1, rad2)
	io.WriteString(g.buf, s)
}

func (g *ATT) Neg(rad1 string) {
	s := fmt.Sprintf("  neg %%%s\n", rad1)
	io.WriteString(g.buf, s)
}

func (g *ATT) Ret() {
	io.WriteString(g.buf, "  ret\n")
}

func (g *ATT) Label(name string) {
	s := fmt.Sprintf("%s:\n", name)
	io.WriteString(g.buf, s)
}

func (g *ATT) Globl(label string) {
	g.Text(fmt.Sprintf(".globl %s", label))
}

func (g *ATT) String(value string) {
	g.Text(fmt.Sprintf(".string \"%s\"", value))
}

func (g *ATT) Size(value int) {
	g.Text(fmt.Sprintf(".size, %d", value))
}

func (g *ATT) Data() {
	g.Text(".data")
}

func (g *ATT) Text(text string) {
	s := fmt.Sprintf("  %s\n", text)
	io.WriteString(g.buf, s)
}

func (g *ATT) Address(name string) string {
	return fmt.Sprintf("(%%%s)", name)
}

func (g *ATT) Index(base, unit string, size int) string {
	return fmt.Sprintf("%s,%%%s,%d", base, unit, size)
}

func prefixed(src string) string {
	_, err := strconv.Atoi(src)
	if err == nil {
		return fmt.Sprintf("$%s", src)
	}

	// address
	if src[0] == '(' {
		return src
	}

	return fmt.Sprintf("%%%s", src)
}

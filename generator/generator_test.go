package generator

import (
	"bytes"
	"go9cc/parser"
	"go9cc/token"
	"testing"
)

func TestGenerator(t *testing.T) {
	input := " (5 + 5) * 5 / 2"
	want := `.intel_syntax noprefix
.globl main
main:
  push 5
  push 5
  pop rdi
  pop rax
  add rax, rdi
  push rax
  push 5
  pop rdi
  pop rax
  imul rax, rdi
  push rax
  push 2
  pop rdi
  pop rax
  cqo
  idiv rdi
  push rax
  pop rax
  ret
`

	out := bytes.NewBufferString("")
	tzer := token.New(input)
	p := parser.New(tzer)
	g := New(p, out)
	g.Gen()

	if out.String() != want {
		t.Fatalf("Wrong generated code =======\nGot:\n%s\n\nWant:\n%s\n", out.String(), want)
	}
}

package generator

import (
	"bytes"
	"go9cc/parser"
	"go9cc/token"
	"testing"
)

func TestGenerator(t *testing.T) {
	want1 := `.intel_syntax noprefix
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
	want2 := `.intel_syntax noprefix
.globl main
main:
  push 5
  push 5
  pop rdi
  pop rax
  imul rax, rdi
  push rax
  push 5
  push 2
  pop rdi
  pop rax
  imul rax, rdi
  push rax
  pop rdi
  pop rax
  cmp rax, rdi
  sete al
  movzb rax, al
  push rax
  pop rax
  ret
`

	tests := []struct {
		input string
		want  string
	}{
		{" (5 + 5) * 5 / 2", want1},
		{"(5 * 5) == (5 * 2)", want2},
	}

	for i, tt := range tests {
		out := bytes.NewBufferString("")
		tzer := token.New(tt.input)
		p := parser.New(tzer)
		g := New(p, out)
		g.Gen()

		if out.String() != tt.want {
			t.Fatalf(
				"%d: Wrong generated code =======\nGot:\n%s\n\nWant:\n%s\n",
				i,
				out.String(),
				tt.want,
			)
		}
	}
}

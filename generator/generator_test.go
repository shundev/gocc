package generator

import (
	"bytes"
	"go9cc/parser"
	"go9cc/token"
	"testing"
)

func TestGenerator(t *testing.T) {
	want := `.intel_syntax noprefix
.globl main
main:
  push rbp
  mov rbp, rsp
  sub rsp, 0
  mov rsp, rbp
  pop rbp
  ret
`
	tests := []struct {
		input string
		want  string
	}{
		{"", want},
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

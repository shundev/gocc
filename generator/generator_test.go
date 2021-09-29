package generator

import (
	"bytes"
	"go9cc/parser"
	"go9cc/token"
	"testing"
)

func TestGenerator(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{}

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

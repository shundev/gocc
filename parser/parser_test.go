package parser

import (
	"go9cc/token"
	"testing"
)

func TestParseInfix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"-1 + (10 * -2) - 5 / 100;",
			"(((-1) + (10 * (-2))) - (5 / 100));",
		},
		{
			"10 + 5 == 5 * 3;",
			"((10 + 5) == (5 * 3));",
		},
		{
			"(10 == 4) == (3 == 2);",
			"((10 == 4) == (3 == 2));",
		},
		{
			"10 + 5 != 5 * 3;",
			"((10 + 5) != (5 * 3));",
		},
		{
			"10 < 5 == 1 > 3;",
			"((10 < 5) == (1 > 3));",
		},
		{
			"10 < (5 == 1);",
			"(10 < (5 == 1));",
		},
		{
			"10 <= 5 == 1 >= 3;",
			"((10 <= 5) == (1 >= 3));",
		},
		{
			"int ab1000 = 999;",
			"int ab1000 = 999;",
		},
		{
			"int a = 1; int b = 1; int c = 1; a = b = c = 1;",
			"int a = 1; int b = 1; int c = 1; (a = (b = (c = 1)));",
		},
		{
			"",
			"",
		},
		{
			"(1 + 2) == (5 - 2);",
			"((1 + 2) == (5 - 2));",
		},
		{
			"int a = 10;int b = 10; int c = 20;return a + b + c;",
			"int a = 10; int b = 10; int c = 20; return ((a + b) + c);",
		},
		{
			"int a = 10;return a; return 20;",
			"int a = 10; return a; return 20;",
		},
		{
			"if (a == 10) return b;",
			"if ((a == 10)) return b;",
		},
		{
			"int a = 0; if (a = 1 == 10) return b; else return a + 10;",
			"int a = 0; if ((a = (1 == 10))) return b; else return (a + 10);",
		},
		{
			"while (a == 10) return a;",
			"while ((a == 10)) return a;",
		},
		{
			"int a = 10; for (int i=0; i<10;i = i + 1) a = a + 3;",
			"int a = 10; for (int i = 0;(i < 10);(i = (i + 1))) (a = (a + 3));",
		},
		{
			"int i = 0; for (; i<10;) i = i + 1;",
			"int i = 0; for (;(i < 10);) (i = (i + 1));",
		},
		{
			"if (1) { a; b; c; return d;}",
			"if (1) { a; b; c; return d; }",
		},
		{
			"foo    ( );",
			"foo ();",
		},
		{
			"--a;",
			"(-(-a));",
		},
		{
			"&*a;",
			"(&(*a));",
		},
		{
			"*(&a-1);",
			"(*((&a) - 1));",
		},
		{
			"int a = 0;",
			"int a = 0;",
		},
	}

	for i, tt := range tests {
		tzer := token.New(tt.input)
		p := New(tzer)
		node := p.Parse()
		testNode(t, i, node, tt.want)
	}

}

func testNode(t *testing.T, i int, node Node, want string) {
	if node == nil {
		t.Fatalf("%d: Node is nil: want=%s", i, want)
	}

	if node.String() != want {
		t.Fatalf("%d: Wrong Node: got=%s, want=%s", i, node.String(), want)
	}
}

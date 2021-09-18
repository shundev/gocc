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
			"ab1000 = 999;",
			"(ab1000 = 999);",
		},
		{
			"a = b = c = 1;",
			"(a = (b = (c = 1)));",
		},
		{
			"a = b = 1; a + b;",
			"(a = (b = 1)); (a + b);",
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
			"a = 10;b = c = 20;return a + b + c;",
			"(a = 10); (b = (c = 20)); return ((a + b) + c);",
		},
		{
			"a = 10;return a; return 20;",
			"(a = 10); return a; return 20;",
		},
		{
			"if (a == 10) return b;",
			"if ((a == 10)) { return b; }",
		},
		{
			"if (a = 1 == 10) return b; else return a + 10;",
			"if ((a = (1 == 10))) { return b; } else { return (a + 10); }",
		},
		{
			"while (a == 10) return a;",
			"while ((a == 10)) { return a; }",
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

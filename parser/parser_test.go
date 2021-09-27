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
			"int main() { int a; return a;}",
			"int main () { int a; return a; }",
		},
		{
			"int main () { int *a = 0, **b, ***c; }",
			"int main () { int* a = 0; int** b, *** c; }",
		},
		{
			"int main() {-1 + (10 * -2) - 5 / 100;}",
			"int main () { (((-1) + (10 * (-2))) - (5 / 100)); }",
		},
		{
			"int main() { 10 + 5 == 5 * 3; }",
			"int main () { ((10 + 5) == (5 * 3)); }",
		},
		{
			"int main () { (10 == 4) == (3 == 2); }",
			"int main () { ((10 == 4) == (3 == 2)); }",
		},
		{
			"int main () { 10 + 5 != 5 * 3; }",
			"int main () { ((10 + 5) != (5 * 3)); }",
		},
		{
			"int main () { 10 < 5 == 1 > 3; }",
			"int main () { ((10 < 5) == (1 > 3)); }",
		},
		{
			"int main () { 10 < (5 == 1); }",
			"int main () { (10 < (5 == 1)); }",
		},
		{
			"int main () { 10 <= 5 == 1 >= 3; }",
			"int main () { ((10 <= 5) == (1 >= 3)); }",
		},
		{
			"int main () { int ab1000 = 999; }",
			"int main () { int ab1000 = 999; }",
		},
		{
			"int main () { int a = 1; int b = 1; int c = 1; a = b = c = 1; }",
			"int main () { int a = 1; int b = 1; int c = 1; (a = (b = (c = 1))); }",
		},
		{
			"int main () { (1 + 2) == (5 - 2); }",
			"int main () { ((1 + 2) == (5 - 2)); }",
		},
		{
			"int main () { int a = 10;int b = 10; int c = 20;return a + b + c; }",
			"int main () { int a = 10; int b = 10; int c = 20; return ((a + b) + c); }",
		},
		{
			"int main () { int a = 10; return a; return 20; }",
			"int main () { int a = 10; return a; return 20; }",
		},
		{
			"int main () { if (a == 10) return b; }",
			"int main () { if ((a == 10)) return b; }",
		},
		{
			"int main () { int a = 0; if (a = 1 == 10) return b; else return a + 10; }",
			"int main () { int a = 0; if ((a = (1 == 10))) return b; else return (a + 10); }",
		},
		{
			"int main () { while (a == 10) return a; }",
			"int main () { while ((a == 10)) return a; }",
		},
		{
			"int main () { int a = 10; for (int i=0; i<10;i = i + 1) a = a + 3; }",
			"int main () { int a = 10; for (int i = 0;;(i < 10);(i = (i + 1))) (a = (a + 3)); }",
		},
		{
			"int main () { int i = 0; for (; i<10;) i = i + 1; }",
			"int main () { int i = 0; for (;(i < 10);) (i = (i + 1)); }",
		},
		{
			"int main () { if (1) { a; b; c; return d;} }",
			"int main () { if (1) { a; b; c; return d; } }",
		},
		{
			"int main () { foo    ( ); }",
			"int main () { foo (); }",
		},
		{
			"int main () { --a; }",
			"int main () { (-(-a)); }",
		},
		{
			"int main () { &*a; }",
			"int main () { (&(*a)); }",
		},
		{
			"int main () { *(&a-1); }",
			"int main () { (*((&a) - 1)); }",
		},
		{
			"int main () { int a = 0; }",
			"int main () { int a = 0; }",
		},
		{
			"int main () { int; }",
			"int main () {  }",
		},
		{
			"int main () { int a; }",
			"int main () { int a; }",
		},
		{
			"int foo (int a, int b, int hello99) { return a + b + hello99; }",
			"int foo (int a, int b, int hello99) { return ((a + b) + hello99); }",
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

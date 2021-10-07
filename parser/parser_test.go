package parser

import (
	"go9cc/ast"
	"go9cc/token"
	"testing"
)

func TestParseInfix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"char a[6] = \"hello\"; int main() { return 0; }",
			"char[6] a = \"hello\"; int main () { return 0; }",
		},
		{
			"char a; int main() { char x = 1; return 0; }",
			"char a; int main () { char x = 1; return 0; }",
		},
		{
			"int x; int* y; int main() { x = 10; return x; }",
			"int x; int* y; int main () { (x = 10); return x; }",
		},
		{
			"int main() { int a[10]; a[5] = 10; a[4] = 5; return a[4] + a[5];}",
			"int main () { int[10] a; (a[5] = 10); (a[4] = 5); return (a[4] + a[5]); }",
		},
		{
			"int main() { int a[10]; return *a;}",
			"int main () { int[10] a; return (*a); }",
		},
		{
			"int main() { int a; return a;}",
			"int main () { int a; return a; }",
		},
		{
			"int main () { int *a, **b, ***c; }",
			"int main () { int* a, int** b, int*** c; }",
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
			"int main () { int a,b; if (a == 10) return b; }",
			"int main () { int a, int b; if ((a == 10)) return b; }",
		},
		{
			"int main () { int a,b = 0; if (a = 1 == 10) return b; else return a + 10; }",
			"int main () { int a, int b = 0; if ((a = (1 == 10))) return b; else return (a + 10); }",
		},
		{
			"int main () { int a; while (a == 10) return a; }",
			"int main () { int a; while ((a == 10)) return a; }",
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
			"int main () { int a,b,c,d; if (1) { a; b; c; return d;} }",
			"int main () { int a, int b, int c, int d; if (1) { a; b; c; return d; } }",
		},
		{
			"int foo() { } int main () { foo    ( ); }",
			"int foo () {  } int main () { foo(); }",
		},
		{
			"int main () { int a; --a; }",
			"int main () { int a; (-(-a)); }",
		},
		{
			"int main () { int a; &*a; }",
			"int main () { int a; (&(*a)); }",
		},
		{
			"int main () { int a; *(&a-1); }",
			"int main () { int a; (*((&a) - 1)); }",
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
		{
			"int foo (int a, int b, int hello99) { return a + b + hello99; } int main() { return foo(1,2,3); }",
			"int foo (int a, int b, int hello99) { return ((a + b) + hello99); } int main () { return foo(1, 2, 3); }",
		},
		{
			"int main() { int x; sizeof(x + 4); }",
			"int main () { int x; (sizeof(x + 4)); }",
		},
		{
			"int main() { int x; sizeof x * 4; }",
			"int main () { int x; ((sizeofx) * 4); }",
		},
	}

	for i, tt := range tests {
		tzer := token.New(tt.input)
		p := New(tzer)
		node := p.Parse()
		testNode(t, i, node, tt.want)
	}

}

func testNode(t *testing.T, i int, node ast.Node, want string) {
	if node == nil {
		t.Fatalf("%d: Node is nil: want=%s", i, want)
	}

	if node.String() != want {
		t.Fatalf("%d: Wrong Node:\ngot =%s,\nwant=%s", i, node.String(), want)
	}
}

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
			"-1 + (10 * -2) - 5 / 100",
			"(((-1) + (10 * (-2))) - (5 / 100))",
		},
		{
			"10 + 5 == 5 * 3",
			"((10 + 5) == (5 * 3))",
		},
		{
			"(10 == 4) == (3 == 2)",
			"((10 == 4) == (3 == 2))",
		},
		{
			"10 + 5 != 5 * 3",
			"((10 + 5) != (5 * 3))",
		},
		{
			"10 < 5 == 1 > 3",
			"((10 < 5) == (1 > 3))",
		},
		{
			"10 < (5 == 1)",
			"(10 < (5 == 1))",
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

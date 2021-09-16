package parser

import (
	"go9cc/token"
	"testing"
)

func TestParseInfix(t *testing.T) {
	/*
		            -
		      +           /
			 1      *    5     100
			     10   2
	*/
	input := "1 + (10 * 2) - 5 / 100"
	tzer := token.New(input)
	p := New(tzer)
	node := p.Parse()
	testNode(t, node, "((1 + (10 * 2)) - (5 / 100))")
}

func testNode(t *testing.T, node Node, want string) {
	if node == nil {
		t.Fatalf("Node is nil: want=%s", want)
	}

	if node.String() != want {
		t.Fatalf("Wrong Node: got=%s, want=%s", node.String(), want)
	}

}

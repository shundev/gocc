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
	testNode(t, node, ND_INFIX, 0, "-")
	testNode(t, node.Left, ND_INFIX, 0, "+")
	testNode(t, node.Left.Left, ND_NUM, 1, "1")
	testNode(t, node.Left.Right, ND_INFIX, 0, "*")
	testNode(t, node.Left.Right.Left, ND_NUM, 10, "10")
	testNode(t, node.Left.Right.Right, ND_NUM, 2, "2")

	testNode(t, node.Right, ND_INFIX, 0, "/")
	testNode(t, node.Right.Left, ND_NUM, 5, "5")
	testNode(t, node.Right.Right, ND_NUM, 100, "100")

	testIsNil(t, node.Left.Left.Left)
	testIsNil(t, node.Right.Right.Right)
}

func testIsNil(t *testing.T, a *Node) {
	if a != nil {
		t.Fatalf("Node is not nil: got=%+v", a)
	}
}

func testNode(t *testing.T, node *Node, kind NodeKind, val int, str string) {
	if node == nil {
		t.Fatalf("Node is nil: want=%s", kind)
	}

	if node.Kind != kind {
		t.Fatalf("Wrong Node.Kind: %s != %s", node.Kind, kind)
	}

	if node.Val != val {
		t.Fatalf("Wrong Node.Val: %d != %d", node.Val, val)
	}

	if node.Str != str {
		t.Fatalf("Wrong Node.Str: %s != %s", node.Str, str)
	}
}

package token

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	input := "()10+-333333     *400/)==!="
	tzer := New(input)
	cur := tzer.Tokenize()

	testToken(t, cur, LPAREN, 0, "(", 0)
	cur = cur.Next
	testToken(t, cur, RPAREN, 0, ")", 1)
	cur = cur.Next
	testToken(t, cur, NUM, 10, "10", 2)
	cur = cur.Next
	testToken(t, cur, PLUS, 0, "+", 4)
	cur = cur.Next
	testToken(t, cur, MINUS, 0, "-", 5)
	cur = cur.Next
	testToken(t, cur, NUM, 333333, "333333", 6)
	cur = cur.Next
	testToken(t, cur, ASTERISK, 0, "*", 17)
	cur = cur.Next
	testToken(t, cur, NUM, 400, "400", 18)
	cur = cur.Next
	testToken(t, cur, SLASH, 0, "/", 21)
	cur = cur.Next
	testToken(t, cur, RPAREN, 0, ")", 22)
	cur = cur.Next
	testToken(t, cur, EQ, 0, "==", 23)
	cur = cur.Next
	testToken(t, cur, EQ, 0, "!=", 25)
	cur = cur.Next
	testToken(t, cur, EOF, 0, "", 27)
}

func testToken(t *testing.T, token *Token, kind TokenKind, val int, str string, col int) {
	if token.Kind != kind {
		t.Errorf("Wrong TokenKind: %s != %s", token.Kind, kind)
	}

	if token.Val != val {
		t.Errorf("Wrong Token.Val: %d != %d", token.Val, val)
	}

	if token.Str != str {
		t.Errorf("Wrong Token.Str: %s != %s", token.Str, str)
	}
	if token.Col != col {
		t.Errorf("Wrong Token.Col: %d != %d", token.Col, col)
	}

}

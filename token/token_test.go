package token

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	input := "()10+-333333     *400/)==!=<><=>=a100=z かなカナ漢字 🍺;returna return*ABC_Z _H if else while do{}for&int**,sizeof"
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
	testToken(t, cur, NEQ, 0, "!=", 25)
	cur = cur.Next
	testToken(t, cur, LT, 0, "<", 27)
	cur = cur.Next
	testToken(t, cur, GT, 0, ">", 28)
	cur = cur.Next
	testToken(t, cur, LTE, 0, "<=", 29)
	cur = cur.Next
	testToken(t, cur, GTE, 0, ">=", 31)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "a100", 33)
	cur = cur.Next
	testToken(t, cur, ASSIGN, 0, "=", 37)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "z", 38)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "かなカナ漢字", 40)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "🍺", 47)
	cur = cur.Next
	testToken(t, cur, SEMICOLLON, 0, ";", 48)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "returna", 49)
	cur = cur.Next
	testToken(t, cur, RETURN, 0, "return", 57)
	cur = cur.Next
	testToken(t, cur, ASTERISK, 0, "*", 63)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "ABC_Z", 64)
	cur = cur.Next
	testToken(t, cur, IDENT, 0, "_H", 70)
	cur = cur.Next
	testToken(t, cur, IF, 0, "if", 73)
	cur = cur.Next
	testToken(t, cur, ELSE, 0, "else", 76)
	cur = cur.Next
	testToken(t, cur, WHILE, 0, "while", 81)
	cur = cur.Next
	testToken(t, cur, DO, 0, "do", 87)
	cur = cur.Next
	testToken(t, cur, LBRACE, 0, "{", 89)
	cur = cur.Next
	testToken(t, cur, RBRACE, 0, "}", 90)
	cur = cur.Next
	testToken(t, cur, FOR, 0, "for", 91)
	cur = cur.Next
	testToken(t, cur, AND, 0, "&", 94)
	cur = cur.Next
	testToken(t, cur, TYPE, 0, "int", 95)
	cur = cur.Next
	testToken(t, cur, ASTERISK, 0, "*", 98)
	cur = cur.Next
	testToken(t, cur, ASTERISK, 0, "*", 99)
	cur = cur.Next
	testToken(t, cur, COMMA, 0, ",", 100)
	cur = cur.Next
	testToken(t, cur, SIZEOF, 0, "sizeof", 101)
	cur = cur.Next
	testToken(t, cur, EOF, 0, "", 107)
}

func testToken(t *testing.T, token *Token, kind TokenKind, val int, str string, col int) {
	if token.Kind != kind {
		t.Fatalf("Wrong TokenKind: %s != %s", token.Kind, kind)
	}

	if token.Val != val {
		t.Fatalf("Wrong Token.Val: %d != %d", token.Val, val)
	}

	if token.Str != str {
		t.Fatalf("Wrong Token.Str: %s != %s", token.Str, str)
	}
	if token.Col != col {
		t.Fatalf("Wrong Token.Col: %d != %d", token.Col, col)
	}

}

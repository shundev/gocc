package token

import (
	"fmt"
	"go9cc/ffmt"
	"os"
)

const (
	RESERVED = "RESERVED"
	NUM      = "NUM"
	EOF      = "EOF"
	START    = "START"
)

type TokenKind string

type Token struct {
	Kind TokenKind
	Next *Token
	Val  int
	Str  string
}

func New(kind TokenKind, curToken *Token, val int, str string) *Token {
	token := Token{Kind: kind, Val: val, Str: str}
	curToken.Next = &token
	return &token
}

type Tokenizer struct {
	idx  int
	code string
}

func (t *Tokenizer) curCh() byte {
	if t.idx >= len(t.code) {
		return 0
	}

	return t.code[t.idx]
}

func (t *Tokenizer) Tokenize(code string) *Token {
	t.code = code
	t.idx = skip(code, 0)

	head := &Token{START, nil, 0, ""}
	cur := head

	for {
		switch t.curCh() {
		case '+':
			cur = New(RESERVED, cur, 0, "+")
			t.idx++
		case '-':
			cur = New(RESERVED, cur, 0, "-")
			t.idx++
		case 0:
			token := New(EOF, cur, 0, "")
			cur = token
			return head.Next
		default:
			if isDigit(t.curCh()) {
				intVal, newIdx := readInteger(t.code, t.idx)
				cur = New(NUM, cur, intVal, fmt.Sprintf("%d", intVal))
				t.idx = newIdx
			} else {
				ffmt.Err("Unexpected char: %s", string(t.curCh()))
				os.Exit(1)
			}
		}

		t.idx = skip(code, t.idx)
	}
}

func isWS(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func skip(s string, start int) int {
	p := start
	for p < len(s) && isWS(s[p]) {
		p++
	}

	return p
}

func readInteger(s string, start int) (int, int) {
	p := skip(s, start)
	val := 0
	for p < len(s) && isDigit(s[p]) {
		val *= 10
		val += int(s[p] - 48)
		p++
	}

	return val, p
}

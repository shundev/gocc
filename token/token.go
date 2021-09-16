package token

import (
	"fmt"
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
	Col  int
}

func New(kind TokenKind, curToken *Token, val int, str string, col int) *Token {
	token := Token{Kind: kind, Val: val, Str: str, Col: col}
	curToken.Next = &token
	return &token
}

type Tokenizer struct {
	idx  int
	code []rune
}

func (t *Tokenizer) Error(pos int, msg string, args ...interface{}) {
	fmt.Println(pos)
	fmt.Fprintln(os.Stderr, string(t.code))
	for i := 0; i < pos; i++ {
		fmt.Printf(" ")
	}
	fmt.Fprintf(os.Stderr, "^ "+msg+"\n", args...)
	os.Exit(1)
}

func (t *Tokenizer) curCh() rune {
	if t.idx >= len(t.code) {
		return 0
	}

	return t.code[t.idx]
}

func (t *Tokenizer) Tokenize(code string) *Token {
	t.code = []rune(code)
	t.idx = skip(t.code, 0)

	head := &Token{START, nil, 0, "", 0}
	cur := head

	for {
		switch t.curCh() {
		case '+':
			cur = New(RESERVED, cur, 0, "+", t.idx)
			t.idx++
		case '-':
			cur = New(RESERVED, cur, 0, "-", t.idx)
			t.idx++
		case 0:
			token := New(EOF, cur, 0, "", t.idx)
			cur = token
			return head.Next
		default:
			if isDigit(t.curCh()) {
				intVal, newIdx := readInteger(t.code, t.idx)
				cur = New(NUM, cur, intVal, fmt.Sprintf("%d", intVal), t.idx)
				t.idx = newIdx
			} else {
				t.idx = skip(t.code, t.idx)
				t.Error(t.idx, "Unexpected char: %s", string(t.curCh()))
				os.Exit(1)
			}
		}

		t.idx = skip(t.code, t.idx)
	}
}

func isWS(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func skip(s []rune, start int) int {
	p := start
	for p < len(s) && isWS(s[p]) {
		p++
	}

	return p
}

func readInteger(s []rune, start int) (int, int) {
	p := skip(s, start)
	val := 0
	for p < len(s) && isDigit(s[p]) {
		val *= 10
		val += int(s[p] - 48)
		p++
	}

	return val, p
}

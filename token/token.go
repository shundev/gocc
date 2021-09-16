package token

import (
	"fmt"
	"os"
	"strings"
)

const (
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	LPAREN   = "("
	RPAREN   = ")"
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

func newToken(kind TokenKind, curToken *Token, val int, str string, col int) *Token {
	token := Token{Kind: kind, Val: val, Str: str, Col: col}
	curToken.Next = &token
	return &token
}

type Tokenizer struct {
	idx  int
	code []rune
}

func New(code string) *Tokenizer {
	return &Tokenizer{idx: 0, code: []rune(code)}
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

func (t *Tokenizer) Expect(token *Token, kinds ...TokenKind) {
	match := false
	ss := []string{}
	for _, kind := range kinds {
		match = match || token.Kind == kind
		ss = append(ss, string(kind))
	}

	if !match {
		t.Error(token.Col, "Expected %s. Got %s.", strings.Join(ss, " or "), token.Kind)
		os.Exit(1)
	}
}

func (t *Tokenizer) curCh() rune {
	if t.idx >= len(t.code) {
		return 0
	}

	return t.code[t.idx]
}

func (t *Tokenizer) Tokenize() *Token {
	t.idx = skip(t.code, 0)

	head := &Token{START, nil, 0, "", 0}
	cur := head

	for {
		switch t.curCh() {
		case '+':
			cur = newToken(PLUS, cur, 0, string(t.curCh()), t.idx)
			t.idx++
		case '-':
			cur = newToken(MINUS, cur, 0, string(t.curCh()), t.idx)
			t.idx++
		case '*':
			cur = newToken(ASTERISK, cur, 0, string(t.curCh()), t.idx)
			t.idx++
		case '/':
			cur = newToken(SLASH, cur, 0, string(t.curCh()), t.idx)
			t.idx++
		case '(':
			cur = newToken(LPAREN, cur, 0, string(t.curCh()), t.idx)
			t.idx++
		case ')':
			cur = newToken(RPAREN, cur, 0, string(t.curCh()), t.idx)
			t.idx++
		case 0:
			token := newToken(EOF, cur, 0, "", t.idx)
			cur = token
			return head.Next
		default:
			if isDigit(t.curCh()) {
				intVal, newIdx := readInteger(t.code, t.idx)
				cur = newToken(NUM, cur, intVal, fmt.Sprintf("%d", intVal), t.idx)
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

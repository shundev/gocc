package token

import (
	"bytes"
	"fmt"
	"go9cc/emoji"
	"os"
	"strings"
	"unicode"
)

const (
	PLUS       = "+"
	MINUS      = "-"
	ASTERISK   = "*"
	SLASH      = "/"
	LPAREN     = "("
	RPAREN     = ")"
	ASSIGN     = "="
	EQ         = "=="
	NEQ        = "!="
	LT         = "<"
	LTE        = "<="
	GT         = ">"
	GTE        = ">="
	NUM        = "NUM"
	IDENT      = "IDENT"
	SEMICOLLON = ";"
	RETURN     = "RETURN"
	EOF        = "EOF"
	START      = "START"
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
		case '<':
			t.idx++
			if t.curCh() == '=' {
				t.idx--
				cur = newToken(LTE, cur, 0, "<=", t.idx)
				t.idx += 2
			} else {
				t.idx--
				cur = newToken(LT, cur, 0, string(t.curCh()), t.idx)
				t.idx++
			}
		case '>':
			t.idx++
			if t.curCh() == '=' {
				t.idx--
				cur = newToken(GTE, cur, 0, ">=", t.idx)
				t.idx += 2
			} else {
				t.idx--
				cur = newToken(GT, cur, 0, string(t.curCh()), t.idx)
				t.idx++
			}
		case '=':
			t.idx++
			if t.curCh() == '=' {
				t.idx--
				cur = newToken(EQ, cur, 0, "==", t.idx)
				t.idx += 2
			} else {
				t.idx--
				cur = newToken(ASSIGN, cur, 0, "=", t.idx)
				t.idx++
			}
		case '!':
			t.idx++
			if t.curCh() != '=' {
				t.Error(t.idx, "Unexpected char: %s", string(t.curCh()))
			}
			t.idx--
			cur = newToken(NEQ, cur, 0, "!=", t.idx)
			t.idx += 2
		case ';':
			cur = newToken(SEMICOLLON, cur, 0, ";", t.idx)
			t.idx++
		case 0:
			cur = newToken(EOF, cur, 0, "", t.idx)
			return head.Next
		default:
			strVal, newIdx := tryKeywords(t.code, t.idx)
			if strVal == "return" {
				cur = newToken(RETURN, cur, 0, strVal, t.idx)
				t.idx = newIdx
			} else if isDigit(t.curCh()) {
				intVal, newIdx := readInteger(t.code, t.idx)
				cur = newToken(NUM, cur, intVal, fmt.Sprintf("%d", intVal), t.idx)
				t.idx = newIdx
			} else if isIdent(t.curCh()) {
				strVal, newIdx := readIdent(t.code, t.idx)
				cur = newToken(IDENT, cur, 0, strVal, t.idx)
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

func isIdentStart(ch rune) bool {
	return 'a' <= ch && ch <= 'z'
}

func isIdent(ch rune) bool {
	if 'a' <= ch && ch <= 'z' {
		return true
	}

	if unicode.In(ch, unicode.Katakana) {
		return true
	}

	if unicode.In(ch, unicode.Hiragana) {
		return true
	}

	if unicode.In(ch, unicode.Han) {
		return true
	}

	if emoji.In(ch) {
		return true
	}

	return false
}

func readIdent(s []rune, start int) (string, int) {
	idx := start
	var out bytes.Buffer
	for idx < len(s) && (isIdent(s[idx]) || isDigit(s[idx])) {
		out.WriteRune(s[idx])
		idx++
	}

	return out.String(), idx
}

func tryKeywords(s []rune, start int) (string, int) {
	// TODO optimize
	ss := string(s[start:])
	if !strings.HasPrefix(ss, "return") {
		return "", start
	}

	if ss == "return" {
		return "return", start + 6
	}

	r := []rune(ss[6:])
	if !isIdent(r[0]) {
		return "return", start + 6
	}

	return "", start
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

func skip(s []rune, start int) int {
	p := start
	for p < len(s) && isWS(s[p]) {
		p++
	}

	return p
}

func equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

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
	LBRACE     = "{"
	RBRACE     = "}"
	LBRACKET   = "["
	RBRACKET   = "]"
	ASSIGN     = "="
	EQ         = "=="
	NEQ        = "!="
	LT         = "<"
	LTE        = "<="
	GT         = ">"
	GTE        = ">="
	AND        = "&"
	NUM        = "NUM"
	IDENT      = "IDENT"
	TYPE       = "TYPE"
	SEMICOLLON = ";"
	COMMA      = ","
	RETURN     = "RETURN"
	SIZEOF     = "SIZEOF"
	IF         = "IF"
	ELSE       = "ELSE"
	WHILE      = "WHILE"
	FOR        = "FOR"
	DO         = "DO"
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
	col  int
	code []rune
}

func New(code string) *Tokenizer {
	return &Tokenizer{col: 0, code: []rune(code)}
}

func (t *Tokenizer) Error(token *Token, msg string, args ...interface{}) {
	fmt.Println(token.Col)
	fmt.Fprintln(os.Stderr, string(t.code))
	for i := 0; i < token.Col; i++ {
		fmt.Printf(" ")
	}
	fmt.Fprintf(os.Stderr, "^ "+msg+"\n", args...)
	os.Exit(1)
}

func (t *Tokenizer) errorCurrent(msg string, args ...interface{}) {
	fmt.Println(t.col)
	fmt.Fprintln(os.Stderr, string(t.code))
	for i := 0; i < t.col; i++ {
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
		t.Error(token, "Expected %s. Got %s.", strings.Join(ss, " or "), token.Kind)
		os.Exit(1)
	}
}

func (t *Tokenizer) curCh() rune {
	if t.col >= len(t.code) {
		return 0
	}

	return t.code[t.col]
}

func (t *Tokenizer) Tokenize() *Token {
	t.col = skip(t.code, 0)

	head := &Token{START, nil, 0, "", 0}
	cur := head

	for {
		switch t.curCh() {
		case '+':
			cur = newToken(PLUS, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '-':
			cur = newToken(MINUS, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '*':
			cur = newToken(ASTERISK, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '/':
			cur = newToken(SLASH, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '(':
			cur = newToken(LPAREN, cur, 0, string(t.curCh()), t.col)
			t.col++
		case ')':
			cur = newToken(RPAREN, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '{':
			cur = newToken(LBRACE, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '}':
			cur = newToken(RBRACE, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '[':
			cur = newToken(LBRACKET, cur, 0, string(t.curCh()), t.col)
			t.col++
		case ']':
			cur = newToken(RBRACKET, cur, 0, string(t.curCh()), t.col)
			t.col++
		case ',':
			cur = newToken(COMMA, cur, 0, string(t.curCh()), t.col)
			t.col++
		case '<':
			t.col++
			if t.curCh() == '=' {
				t.col--
				cur = newToken(LTE, cur, 0, "<=", t.col)
				t.col += 2
			} else {
				t.col--
				cur = newToken(LT, cur, 0, string(t.curCh()), t.col)
				t.col++
			}
		case '>':
			t.col++
			if t.curCh() == '=' {
				t.col--
				cur = newToken(GTE, cur, 0, ">=", t.col)
				t.col += 2
			} else {
				t.col--
				cur = newToken(GT, cur, 0, string(t.curCh()), t.col)
				t.col++
			}
		case '=':
			t.col++
			if t.curCh() == '=' {
				t.col--
				cur = newToken(EQ, cur, 0, "==", t.col)
				t.col += 2
			} else {
				t.col--
				cur = newToken(ASSIGN, cur, 0, "=", t.col)
				t.col++
			}
		case '!':
			t.col++
			if t.curCh() != '=' {
				t.errorCurrent("Unexpected char: %s", string(t.curCh()))
			}
			t.col--
			cur = newToken(NEQ, cur, 0, "!=", t.col)
			t.col += 2
		case ';':
			cur = newToken(SEMICOLLON, cur, 0, ";", t.col)
			t.col++
		case '&':
			cur = newToken(AND, cur, 0, "&", t.col)
			t.col++
		case 0:
			cur = newToken(EOF, cur, 0, "", t.col)
			return head.Next
		default:
			if newcol, ok := tryKeyword(t.code, t.col, "sizeof"); ok {
				cur = newToken(SIZEOF, cur, 0, "sizeof", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "int"); ok {
				cur = newToken(TYPE, cur, 0, "int", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "return"); ok {
				cur = newToken(RETURN, cur, 0, "return", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "if"); ok {
				cur = newToken(IF, cur, 0, "if", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "else"); ok {
				cur = newToken(ELSE, cur, 0, "else", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "for"); ok {
				cur = newToken(FOR, cur, 0, "for", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "while"); ok {
				cur = newToken(WHILE, cur, 0, "while", t.col)
				t.col = newcol
			} else if newcol, ok := tryKeyword(t.code, t.col, "do"); ok {
				cur = newToken(DO, cur, 0, "do", t.col)
				t.col = newcol
			} else if isDigit(t.curCh()) {
				intVal, newcol := readInteger(t.code, t.col)
				cur = newToken(NUM, cur, intVal, fmt.Sprintf("%d", intVal), t.col)
				t.col = newcol
			} else if isIdent(t.curCh()) {
				strVal, newcol := readIdent(t.code, t.col)
				cur = newToken(IDENT, cur, 0, strVal, t.col)
				t.col = newcol
			} else {
				t.col = skip(t.code, t.col)
				t.errorCurrent("Unexpected char: %s", string(t.curCh()))
				os.Exit(1)
			}
		}

		t.col = skip(t.code, t.col)
	}
}

func isWS(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isIdentStart(ch rune) bool {
	if 'a' <= ch && ch <= 'z' {
		return true
	}
	if 'A' <= ch && ch <= 'Z' {
		return true
	}
	if ch == '_' {
		return true
	}
	return false
}

func isIdent(ch rune) bool {
	if isIdentStart(ch) {
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
	col := start
	var out bytes.Buffer
	for col < len(s) && (isIdent(s[col]) || isDigit(s[col])) {
		out.WriteRune(s[col])
		col++
	}

	return out.String(), col
}

func tryKeyword(s []rune, start int, keyword string) (int, bool) {
	// TODO optimize
	ss := string(s[start:])
	if !strings.HasPrefix(ss, keyword) {
		return start, false
	}

	if ss == keyword {
		return start + len([]rune(keyword)), true
	}

	r := []rune(ss[len([]rune(keyword)):])
	if !isIdent(r[0]) {
		return start + len([]rune(keyword)), true
	}

	return start, false
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

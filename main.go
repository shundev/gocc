package main

import (
	"fmt"
	"go9cc/token"
	"log"
	"os"
)

func println(s string, args ...interface{}) {
	fmt.Printf("  "+s+"\n", args...)
}

func expect(t *token.Tokenizer, token *token.Token, kind token.TokenKind) {
	if token.Kind != kind {
		t.Error(token.Col, "Expected %s. Got %s.", kind, token.Kind)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Println("Num of args must be 2.")
	}
	arg := os.Args[1]

	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	tzer := &token.Tokenizer{}
	cur := tzer.Tokenize(arg)

	expect(tzer, cur, token.NUM)
	println("mov rax, %d", cur.Val)
	cur = cur.Next

	for cur.Kind != token.EOF {
		expect(tzer, cur, token.RESERVED)
		switch cur.Kind {
		case token.RESERVED:
			if cur.Str == "+" {
				cur = cur.Next
				expect(tzer, cur, token.NUM)
				println("add rax, %d", cur.Val)
			} else if cur.Str == "-" {
				cur = cur.Next
				expect(tzer, cur, token.NUM)
				println("sub rax, %d", cur.Val)
			} else {
				tzer.Error(cur.Col, "Invalid RESERVED.")
			}
		case token.NUM:
			println("mov rax, %d", cur.Val)
		}

		cur = cur.Next
	}

	fmt.Println("  ret")
}

package main

import (
	"fmt"
	"log"
	"os"
)

func println(s string, args ...interface{}) {
	fmt.Printf("  "+s+"\n", args...)
}

func err(s string, args ...interface{}) {
	log.Printf(s+"\n", args...)
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

func writeMain(code string) {
	idx := skip(code, 0)
	val, idx := readInteger(code, idx)
	println("mov rax, %d", val)
	idx = skip(code, idx)

	for idx < len(code) {
		if code[idx] == '+' {
			idx = skip(code, idx+1)
			val, idx = readInteger(code, idx)
			println("add rax, %d", val)
		} else if code[idx] == '-' {
			idx = skip(code, idx+1)
			val, idx = readInteger(code, idx)
			println("sub rax, %d", val)
		} else {
			err("Unexpected char: %s", string(code[idx]))
			os.Exit(1)
		}

		idx = skip(code, idx)
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
	writeMain(arg)
	fmt.Println("  ret")
}

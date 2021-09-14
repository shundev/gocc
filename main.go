package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	arg := args[0]

	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")
	fmt.Printf("  mov rax, %s\n", arg)
	fmt.Println("  ret")
}

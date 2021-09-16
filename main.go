package main

import (
	"go9cc/generator"
	"go9cc/parser"
	"go9cc/token"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Println("Num of args must be 2.")
	}
	arg := os.Args[1]

	tzer := token.New(arg)
	parser := parser.New(tzer)
	gen := generator.New(parser, os.Stdout)
	gen.Gen()
}

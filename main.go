package main

import (
	"fmt"
	"go9cc/generator"
	"go9cc/parser"
	"go9cc/token"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Num of args must be more than 2.")
	}
	arg := os.Args[1]

	var data string

	if arg == "-c" {
		data = os.Args[2]
	} else {
		file, err := os.Open(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "File %s not found.\n", err.Error())
		}
		defer file.Close()

		dat, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read from file %s.\n", err.Error())
		}
		data = string(dat)
	}

	tzer := token.New(data)
	parser := parser.New(tzer)
	gen := generator.New(parser, os.Stdout)
	gen.Gen()
}

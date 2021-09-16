package main

import (
	"go9cc/repl"
	"os"
)

func main() {
	repl := repl.New()
	repl.Run(os.Stdin, os.Stdout)
}

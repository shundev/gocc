package repl

import (
	"bufio"
	"fmt"
	"go9cc/parser"
	"go9cc/token"
	"io"
)

const PROMPT = ">>> "

type Repl struct {
}

func New() *Repl {
	return &Repl{}
}

func (r *Repl) Run(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		tzer := token.New(line)
		p := parser.New(tzer)
		node := p.Parse()

		io.WriteString(out, node.String())
		io.WriteString(out, "\n")
	}
}

package generator

import (
	"fmt"
	"go9cc/parser"
	"go9cc/types"
	"go9cc/writer"
	"io"
	"os"
)

const INTEL_SYNTAX = false

const (
	RAX = "rax"
	RBP = "rbp" // base pointer
	RSP = "rsp" // stack pointer
	AL  = "al"
	RDI = "rdi" // 1st param
	RSI = "rsi" // 2nd param
	RDX = "rdx" // 3rd param
	RCX = "rcx" // 4th param
	R8D = "r8"  // 5th param
	R9D = "r9"  // 6th param
)

var FUNCCALLREGS = []string{RDI, RSI, RDX, RCX, R8D, R9D}

type Generator struct {
	parser    *parser.Parser
	writer    writer.Writer
	lblCnt    int
	currentFn *parser.FuncDefNode
	fns       map[string]*parser.FuncDefNode
}

func New(p *parser.Parser, out io.Writer) *Generator {
	writer := writer.NewIntelWriter(out)
	gen := &Generator{parser: p, writer: writer, fns: map[string]*parser.FuncDefNode{}}
	return gen
}

func (g *Generator) Gen() {
	g.writer.Header()

	node := g.parser.Parse()

	g.walk(node)
}

// push corresponding address to the top of stack
func (g *Generator) address(fn *parser.FuncDefNode, node interface{}) {
	switch ty := node.(type) {
	case *parser.UnaryExp:
		// Nested unary.
		if ty.Op == "*" {
			g.walk(ty.Right)
			return
		}
	case *parser.LocalVariable:
		offset := g.getOffset(fn, ty)
		g.writer.Lea(RAX, RBP, -offset)
		return
	case *parser.IdentExp:
		offset := g.getOffset(fn, ty)
		g.writer.Lea(RAX, RBP, -offset)
		return
	}

	fmt.Fprintf(os.Stderr, "address must be a ident node, but got: %T\n", node)
	os.Exit(1)
}

func (g *Generator) walk(node parser.Node) {
	switch ty := node.(type) {
	case *parser.ProgramNode:
		for _, stmt := range ty.FuncDefs {
			g.walk(stmt)
		}
	case *parser.ExpStmt:
		g.walk(ty.Exp)
	case *parser.ReturnStmt:
		g.walk(ty.Exp)
		s := fmt.Sprintf(".L.return.%s", g.currentFn.Name)
		g.writer.Jmp(s)
	case *parser.StmtListNode:
		for _, stmt := range ty.Stmts {
			g.walk(stmt)
		}
	case *parser.BlockStmt:
		g.walk(ty.Stmts)
	case *parser.ForStmt:
		stmt, _ := node.(*parser.ForStmt)
		lblBegin := g.genLbl()
		lblEnd := g.genLbl()
		if stmt.Init != nil {
			g.walk(stmt.Init)
		}
		g.writer.Label(lblBegin)
		if stmt.Cond != nil {
			g.walk(stmt.Cond)
			g.writer.Cmp(RAX, "0")
			g.writer.Je(lblEnd) // RAXが0(false)ならforの外にジャンプ
		}
		g.walk(stmt.Body)
		if stmt.AfterEach != nil {
			g.walk(stmt.AfterEach)
		}
		g.writer.Jmp(lblBegin)
		g.writer.Label(lblEnd)
	case *parser.WhileStmt:
		stmt, _ := node.(*parser.WhileStmt)
		lblBegin := g.genLbl()
		lblEnd := g.genLbl()
		g.writer.Label(lblBegin)
		g.walk(stmt.Cond)
		g.writer.Cmp(RAX, "0")
		g.writer.Je(lblEnd) // RAXが0(false)ならwhileの外にジャンプ
		g.walk(stmt.Body)
		g.writer.Jmp(lblBegin)
		g.writer.Label(lblEnd)
	case *parser.IfStmt:
		stmt, _ := node.(*parser.IfStmt)
		lblElse := g.genLbl()
		lblEnd := g.genLbl()
		g.walk(stmt.Cond)
		g.writer.Cmp(RAX, "0")
		g.writer.Je(lblElse) // RAXが0(false)ならelseブロックにジャンプ
		g.walk(stmt.IfBody)
		g.writer.Jmp(lblEnd)
		g.writer.Label(lblElse)
		if stmt.ElseBody != nil {
			g.walk(stmt.ElseBody)
		}
		g.writer.Label(lblEnd)
	case *parser.NumExp:
		val := fmt.Sprintf("%d", ty.Val)
		g.writer.Mov(RAX, val)
	case *parser.IdentExp:
		// 変数呼び出し
		g.address(g.currentFn, ty)
		g.writer.Mov(RAX, g.writer.Address(RAX))
	case *parser.FuncCallExp:
		// 関数呼び出し
		// FIXME: 型チェック
		for i, param := range ty.Params.Exps {
			g.walk(param)
			if i >= len(FUNCCALLREGS) {
				g.writer.Push(RAX)
			} else {
				g.writer.Mov(FUNCCALLREGS[i], RAX)
			}
		}
		g.writer.Mov(RAX, "0")
		// FIXME: need align before call?
		g.writer.Call(ty.Name)
	case *parser.FuncDefNode:
		g.fns[ty.Name] = ty
		g.currentFn = ty
		g.writer.Label(ty.Name)
		g.prolog(ty.StackSize)

		fn, ok := g.fns[ty.Name]
		if !ok {
			fmt.Fprintf(os.Stderr, "Function %s not found.\n", ty.Name)
			os.Exit(1)
		}

		// Prepare params
		for i, local := range ty.Args.LV.Locals {
			offset := g.getOffset(fn, local)
			if i >= len(FUNCCALLREGS) {
				g.writer.Pop(RDI)
				g.writer.Lea(RAX, RBP, -offset)
				g.writer.Mov(g.writer.Address(RAX), RDI)
			} else {
				g.writer.Lea(RAX, RBP, -offset)
				g.writer.Mov(g.writer.Address(RAX), FUNCCALLREGS[i])
			}
		}

		g.walk(ty.Body)
		g.writer.Label(fmt.Sprintf(".L.return.%s", ty.Name))
		g.epilog()
	case *parser.UnaryExp:
		switch ty.Op {
		case "&":
			g.address(g.currentFn, ty.Right) // RAXに目標のアドレスが載る
		case "*":
			g.walk(ty.Right) // RAXに目標のアドレスが載る
			g.writer.Mov(RAX, g.writer.Address(RAX))
		case "+":
			// do nothing ( +5 -> 5)
		case "-":
			g.walk(ty.Right)
			g.writer.Neg(RAX)
		case "sizeof":
			size := reduceSizeof(ty)
			g.writer.Mov(RAX, fmt.Sprintf("%d", size))
		}
	case *parser.DeclarationStmt:
		for _, local := range ty.LV.Locals {
			if ty.Exp != nil {
				g.address(g.currentFn, local)
				g.writer.Push(RAX) // 直近2つのRAXが必要な場合は前のRAXをスタックに退避
				g.walk(ty.Exp)
				g.writer.Pop(RDI)
				g.writer.Mov(g.writer.Address(RDI), RAX)
			}
		}
		// 戻り値はRAXに入っている
	case *parser.InfixExp:
		infix := ty
		if infix.Op == "=" {
			g.address(g.currentFn, infix.Left)
			g.writer.Push(RAX) // 直近2つのRAXが必要な場合は前のRAXをスタックに退避
			g.walk(infix.Right)
			g.writer.Pop(RDI)
			g.writer.Mov(g.writer.Address(RDI), RAX)
			return
		}

		// ポインタ演算はタイプのサイズによってスケールする
		if infix.Op == "+" || infix.Op == "-" {
			switch infix.Left.Type().(type) {
			case *types.IntPointer:
				infix = parser.Scale(infix)
			}
		}

		g.walk(infix.Right) // 先に計算した方がRDIに入るから右辺を先にしないと-の時問題
		g.writer.Push(RAX)
		g.walk(infix.Left)
		g.writer.Pop(RDI)

		switch infix.Op {
		case "+":
			g.writer.Add(RAX, RDI)
		case "-":
			g.writer.Sub(RAX, RDI) // 右辺をRDIに入れているから
		case "*":
			g.writer.Mul(RAX, RDI)
		case "/":
			g.writer.Div(RDI)
		case ">":
			// swap RAX and RDI
			g.writer.Push(RAX)
			g.writer.Mov(RAX, RDI)
			g.writer.Pop(RDI)
			fallthrough
		case "<":
			g.writer.Cmp(RAX, RDI)
			g.writer.Setl(AL)
			g.writer.Movzb(RAX, AL)
		case ">=":
			g.writer.Push(RAX)
			g.writer.Mov(RAX, RDI)
			g.writer.Pop(RDI)
			fallthrough
		case "<=":
			g.writer.Cmp(RAX, RDI)
			g.writer.Setle(AL)
			g.writer.Movzb(RAX, AL)
		case "==":
			g.writer.Cmp(RAX, RDI)
			g.writer.Sete(AL)
			g.writer.Movzb(RAX, AL)
		case "!=":
			g.writer.Cmp(RAX, RDI)
			g.writer.Setne(AL)
			g.writer.Movzb(RAX, AL)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown node: %T, %s\n", node, node.String())
		os.Exit(1)
	}
}

func (g *Generator) prolog(stackSize int) {
	g.writer.Push(RBP)
	g.writer.Mov(RBP, RSP)
	g.writer.Sub(RSP, fmt.Sprintf("%d", stackSize))
}

func (g *Generator) epilog() {
	g.writer.Mov(RSP, RBP)
	g.writer.Pop(RBP)
	g.writer.Ret()
}

func (g *Generator) genLbl() string {
	s := fmt.Sprintf(".L%d", g.lblCnt)
	g.lblCnt++
	return s
}

func (g *Generator) getOffset(fn *parser.FuncDefNode, node interface{}) int {
	var name string
	var ty types.Type

	switch node := node.(type) {
	case *parser.LocalVariable:
		name = node.Name
		ty = node.Type
	case *parser.IdentExp:
		name = node.Name
		ty = node.Type()
	default:
		fmt.Fprintf(os.Stderr, "Invalid node: %T\n", node)
		os.Exit(1)
	}

	offset, ok := fn.Offsets[name]
	//fmt.Fprintf(os.Stderr, "name=%s, offset=%d\n", name, offset)
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid ident name: %s\n", name)
		os.Exit(1)
	}

	return fn.StackSize - offset + ty.StackSize()
}

func reduceSizeof(unary *parser.UnaryExp) int {
	return unary.Right.Type().Size()
}

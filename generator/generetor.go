package generator

import (
	"fmt"
	"go9cc/parser"
	"go9cc/token"
	"go9cc/types"
	"go9cc/writer"
	"io"
	"os"
)

const (
	DEBUG        = true
	INTEL_SYNTAX = true
)

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
	writer := writer.New(out, INTEL_SYNTAX)
	gen := &Generator{parser: p, writer: writer, fns: map[string]*parser.FuncDefNode{}}
	return gen
}

func (g *Generator) Gen() {
	g.writer.Header()

	node := g.parser.Parse()

	g.walk(node)
	g.writer.Commit()
}

func (g *Generator) Error(token *token.Token, msg string, args ...interface{}) {
	g.parser.Error(token, msg, args...)
}

// push corresponding address to the top of stack
func (g *Generator) address(fn *parser.FuncDefNode, node interface{}) {
	debug("addr:\t%T", node)
	switch ty := node.(type) {
	case *parser.IndexExp:
		offset := g.getOffset(fn, ty.Ident)
		g.walk(ty.Index) // 結果の数値がRAXに乗る
		src := fmt.Sprintf("%s+%s*%d", RBP, RAX, ty.Type().StackSize())
		g.writer.Lea(-offset, src, RAX)
		return
	case *parser.UnaryExp:
		// Nested unary.
		debug("Op:\t%s", ty.Op)
		if ty.Op == "*" {
			g.walk(ty.Right)
			return
		}
	case *parser.LocalVariable:
		offset := g.getOffset(fn, ty)
		g.writer.Lea(-offset, RBP, RAX)
		return
	case *parser.IdentExp:
		offset := g.getOffset(fn, ty)
		g.writer.Lea(-offset, RBP, RAX)
		return
	}

	debug("address must be a ident node, but got: %T", node)
	os.Exit(1)
}

func (g *Generator) walk(node parser.Node) {
	debug("walk:\t%T", node)
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
			g.writer.Cmp("0", RAX)
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
		g.writer.Cmp("0", RAX)
		g.writer.Je(lblEnd) // RAXが0(false)ならwhileの外にジャンプ
		g.walk(stmt.Body)
		g.writer.Jmp(lblBegin)
		g.writer.Label(lblEnd)
	case *parser.IfStmt:
		stmt, _ := node.(*parser.IfStmt)
		lblElse := g.genLbl()
		lblEnd := g.genLbl()
		g.walk(stmt.Cond)
		g.writer.Cmp("0", RAX)
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
		g.writer.Mov(val, RAX)
	case *parser.IndexExp:
		g.address(g.currentFn, ty) // 配列のあるインデックスのアドレスがRAXに乗る
		g.writer.Mov(g.writer.Address(RAX), RAX)
	case *parser.IdentExp:
		// 変数呼び出し
		g.address(g.currentFn, ty)
		g.writer.Mov(g.writer.Address(RAX), RAX)
	case *parser.FuncCallExp:
		// 関数呼び出し
		defined := ty.Def != nil
		if !defined {
			// コンパイル時点では定義なしでも、リンクされるので問題無し
		}

		for i, param := range ty.Params.Exps {
			if defined {
				arg := ty.Def.Args.LV.Locals[i]
				if arg.Type.String() != param.Type().String() {
					g.Error(
						param.Token(),
						"Param types do not match for %s. Expected %s, but got %s.",
						arg.Name, arg.Type.String(), param.Type().String())
				}
			}

			g.walk(param)
			if i >= len(FUNCCALLREGS) {
				g.writer.Push(RAX)
			} else {
				g.writer.Mov(RAX, FUNCCALLREGS[i])
			}
		}
		g.writer.Mov("0", RAX)
		// FIXME: need align before call?
		g.writer.Call(ty.Name)
	case *parser.FuncDefNode:
		g.fns[ty.Name] = ty
		g.currentFn = ty
		g.writer.Label(ty.Name)
		g.prolog()

		fn, ok := g.fns[ty.Name]
		if !ok {
			g.Error(ty.Token(), "Function %s not found.\n", ty.Name)
		}

		// Prepare params
		for i, local := range ty.Args.LV.Locals {
			offset := g.getOffset(fn, local)
			if i >= len(FUNCCALLREGS) {
				g.writer.Pop(RDI)
				g.writer.Lea(-offset, RBP, RAX)
				g.writer.Mov(RDI, g.writer.Address(RAX))
			} else {
				g.writer.Lea(-offset, RBP, RAX)
				g.writer.Mov(FUNCCALLREGS[i], g.writer.Address(RAX))
			}
		}

		g.walk(ty.Body)
		g.writer.Label(fmt.Sprintf(".L.return.%s", ty.Name))
		g.epilog()
	case *parser.UnaryExp:
		debug("Op:\t%s", ty.Op)
		switch ty.Op {
		case "&":
			g.address(g.currentFn, ty.Right) // RAXに目標のアドレスが載る
		case "*":
			g.walk(ty.Right) // RAXに目標のアドレスが載る
			g.writer.Mov(g.writer.Address(RAX), RAX)
		case "+":
			// do nothing ( +5 -> 5)
		case "-":
			g.walk(ty.Right)
			g.writer.Neg(RAX)
		case "sizeof":
			size := reduceSizeof(ty)
			g.writer.Mov(fmt.Sprintf("%d", size), RAX)
		}
	case *parser.DeclarationStmt:
		for _, local := range ty.LV.Locals {
			if ty.Exp != nil {
				g.address(g.currentFn, local)
				g.writer.Push(RAX) // 直近2つのRAXが必要な場合は前のRAXをスタックに退避
				g.walk(ty.Exp)
				g.writer.Pop(RDI)
				g.writer.Mov(RAX, g.writer.Address(RDI))
			}
		}
		// 戻り値はRAXに入っている
	case *parser.InfixExp:
		debug("Op:\t %s", ty.Op)
		infix := ty
		if infix.Op == "=" {
			g.address(g.currentFn, infix.Left)
			g.writer.Push(RAX) // 直近2つのRAXが必要な場合は前のRAXをスタックに退避
			g.walk(infix.Right)
			g.writer.Pop(RDI)
			g.writer.Mov(RAX, g.writer.Address(RDI))
			return
		}

		// ポインタ演算はタイプのサイズによってスケールする
		if infix.Op == "+" || infix.Op == "-" {
			switch infix.Left.Type().(type) {
			case *types.IntPointer:
				infix = parser.Scale(infix)
			case *types.Array:
				infix = parser.Scale(infix)
			}
		}

		g.walk(infix.Right) // 先に計算した方がRDIに入るから右辺を先にしないと-の時問題
		g.writer.Push(RAX)
		g.walk(infix.Left)
		g.writer.Pop(RDI)

		switch infix.Op {
		case "+":
			g.writer.Add(RDI, RAX)
		case "-":
			g.writer.Sub(RDI, RAX) // 右辺をRDIに入れているから
		case "*":
			g.writer.Mul(RDI, RAX)
		case "/":
			g.writer.Div(RDI)
		case ">":
			// swap RAX and RDI
			g.writer.Push(RAX)
			g.writer.Mov(RDI, RAX)
			g.writer.Pop(RDI)
			fallthrough
		case "<":
			g.writer.Cmp(RDI, RAX)
			g.writer.Setl(AL)
			g.writer.Movzb(AL, RAX)
		case ">=":
			g.writer.Push(RAX)
			g.writer.Mov(RDI, RAX)
			g.writer.Pop(RDI)
			fallthrough
		case "<=":
			g.writer.Cmp(RDI, RAX)
			g.writer.Setle(AL)
			g.writer.Movzb(AL, RAX)
		case "==":
			g.writer.Cmp(RDI, RAX)
			g.writer.Sete(AL)
			g.writer.Movzb(AL, RAX)
		case "!=":
			g.writer.Cmp(RDI, RAX)
			g.writer.Setne(AL)
			g.writer.Movzb(AL, RAX)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown node: %T, %s\n", node, node.String())
		os.Exit(1)
	}
}

func (g *Generator) prolog() {
	g.writer.Push(RBP)
	g.writer.Mov(RSP, RBP)
	g.writer.Sub(fmt.Sprintf("%d", g.currentFn.StackSize), RSP)
}

func (g *Generator) epilog() {
	g.writer.Mov(RBP, RSP)
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

	switch node := node.(type) {
	case *parser.LocalVariable:
		name = node.Name
	case *parser.IdentExp:
		name = node.Name
	default:
		err("Invalid node: %T\n", node)
		os.Exit(1)
	}

	offset, ok := fn.Offsets[name]
	if !ok {
		err("Invalid ident name: %s\n", name)
		os.Exit(1)
	}

	debug("name: %s, offset; %d", name, offset)
	return offset
}

func reduceSizeof(unary *parser.UnaryExp) int {
	return unary.Right.Type().Size()
}

func debug(s string, args ...interface{}) {
	if DEBUG {
		err(s, args...)
	}
}

func err(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

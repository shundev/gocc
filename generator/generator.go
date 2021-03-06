package generator

import (
	"fmt"
	"go9cc/ast"
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
	RIP = "rip"
	RAX = "rax"
	RBP = "rbp" // base pointer
	RSP = "rsp" // stack pointer
	RDI = "rdi" // 1st param
	RSI = "rsi" // 2nd param
	RDX = "rdx" // 3rd param
	RCX = "rcx" // 4th param
	R8  = "r8"  // 5th param
	R9  = "r9"  // 6th param

	EAX = "eax"
	EDI = "edi" // 1st param
	ESI = "esi" // 2nd param
	EDX = "edx" // 3rd param
	ECX = "ecx" // 4th param
	R8D = "r8d" // 5th param
	R9D = "r9d" // 6th param

	AL  = "al"
	DIL = "dil"
	SIL = "sil"
	DL  = "dl"
	CL  = "cl"
	R8B = "r8b"
	R9B = "r9b"
)

var FUNCCALLREGS = []string{RDI, RSI, RDX, RCX, R8, R9}

var DWORD = map[string]string{
	RAX: EAX,
	RDI: EDI,
	RSI: ESI,
	RDX: EDX,
	RCX: ECX,
	R8:  R8D,
	R9:  R9D,
}

var BYTE = map[string]string{
	RAX: AL,
	RDI: DIL,
	RSI: SIL,
	RDX: DL,
	RCX: CL,
	R8:  R8B,
	R9:  R9B,
}

type Generator struct {
	parser    *parser.Parser
	writer    writer.Writer
	lblCnt    int
	currentFn *ast.FuncDefNode
	fns       map[string]*ast.FuncDefNode
}

func New(p *parser.Parser, out io.Writer) *Generator {
	writer := writer.New(out, INTEL_SYNTAX)
	gen := &Generator{
		parser: p,
		writer: writer,
		fns:    map[string]*ast.FuncDefNode{},
	}
	return gen
}

func (g *Generator) globals() map[string]*ast.LocalVariable {
	return g.parser.Globals
}

func (g *Generator) strings() []*ast.StringLiteralExp {
	return g.parser.Strings
}

func (g *Generator) Gen() {
	g.writer.Header()

	node := g.parser.Parse()

	for _, str := range g.strings() {
		g.strDef(str)
	}
	g.walk(node)
	g.writer.Commit()
}

func (g *Generator) Error(token *token.Token, msg string, args ...interface{}) {
	g.parser.Error(token, msg, args...)
}

// push corresponding address to the top of stack
func (g *Generator) address(fn *ast.FuncDefNode, node interface{}) {
	debug("addr:\t%T", node)
	switch ty := node.(type) {
	case *ast.IndexExp:
		offset, base := g.getOffset(fn, ty.Ident)
		g.walk(ty.Index) // ??????????????????RAX?????????
		src := g.writer.Index(base, RAX, ty.Type().StackSize())
		g.writer.Lea(offset, src, RAX)
		return
	case *ast.UnaryExp:
		// Nested unary.
		debug("Op:\t%s", ty.Op)
		if ty.Op == "*" {
			g.walk(ty.Right)
			return
		}
	case *ast.LocalVariable:
		if ty.IsLocal {
			offset, base := g.getOffset(fn, ty)
			g.writer.Lea(offset, base, RAX)
			return
		} else {
			g.writer.Lea(ty.Name, RIP, RAX)
		}
	case *ast.IdentExp:
		offset, base := g.getOffset(fn, ty)
		g.writer.Lea(offset, base, RAX)
		return
	}

	debug("address must be a ident node, but got: %T", node)
	os.Exit(1)
}

func (g *Generator) global(node *ast.DeclarationStmt) {
	for _, local := range node.LV.Locals {
		g.writer.Globl(local.Name)
		g.writer.Data()

		tyStr := getType(local.Type.Size())

		if node.Exp == nil {
			s := fmt.Sprintf("%s 0", tyStr)
			g.writer.Text(s)
		} else {
			switch ty := g.eval(node.Exp).(type) {
			case *ast.NumExp:
				s := fmt.Sprintf("%s %d", tyStr, ty.Val)
				g.writer.Label(local.Name)
				g.writer.Text(s)
			case *ast.StringLiteralExp:
				s := fmt.Sprintf("%s %s", tyStr, ty.Label)
				g.writer.Label(local.Name)
				g.writer.Text(s)
			case *ast.ArrayLiteral:
				g.writer.Label(local.Name)
				for _, exp := range ty.Exps {
					num, ok := exp.(*ast.NumExp)
					if !ok {
						g.Error(ty.Token(), "Invalid global rvalue.")
					}

					arrTy := local.Type.(*types.Array)
					baseTyStr := getType(arrTy.Base.Size())
					debug("%T", arrTy.Base)
					s := fmt.Sprintf("%s %d", baseTyStr, num.Val)
					g.writer.Text(s)
				}
			case *ast.UnaryExp:
				r, _ := ty.Right.(*ast.IdentExp)
				s := fmt.Sprintf("%s %s", tyStr, r.Name)
				g.writer.Label(local.Name)
				g.writer.Text(s)
			default:
				g.Error(node.Exp.Token(), "Cannot evaluate rvalue of %s:", node.Exp)
			}
		}
	}
}

func (g *Generator) strDef(node *ast.StringLiteralExp) {
	g.writer.Globl(node.Label)
	g.writer.Data()
	g.writer.Size(node.Length())
	g.writer.Label(node.Label)
	g.writer.String(node.Val)
	g.writer.Text(".text")
}

func (g *Generator) walk(node ast.Node) {
	debug("walk:\t%T", node)
	switch ty := node.(type) {
	case *ast.ProgramNode:
		for _, stmt := range ty.GlobalStmts {
			g.global(stmt)
		}
		for _, stmt := range ty.FuncDefs {
			g.walk(stmt)
		}
	case *ast.ExpStmt:
		g.walk(ty.Exp)
	case *ast.ReturnStmt:
		g.walk(ty.Exp)
		s := fmt.Sprintf(".L.return.%s", g.currentFn.Name)
		g.writer.Jmp(s)
	case *ast.StmtListNode:
		for _, stmt := range ty.Stmts {
			g.walk(stmt)
		}
	case *ast.BlockStmt:
		g.walk(ty.Stmts)
	case *ast.ForStmt:
		stmt, _ := node.(*ast.ForStmt)
		lblBegin := g.genLbl()
		lblEnd := g.genLbl()
		if stmt.Init != nil {
			g.walk(stmt.Init)
		}
		g.writer.Label(lblBegin)
		if stmt.Cond != nil {
			g.walk(stmt.Cond)
			g.writer.Cmp("0", RAX)
			g.writer.Je(lblEnd) // RAX???0(false)??????for?????????????????????
		}
		g.walk(stmt.Body)
		if stmt.AfterEach != nil {
			g.walk(stmt.AfterEach)
		}
		g.writer.Jmp(lblBegin)
		g.writer.Label(lblEnd)
	case *ast.WhileStmt:
		stmt, _ := node.(*ast.WhileStmt)
		lblBegin := g.genLbl()
		lblEnd := g.genLbl()
		g.writer.Label(lblBegin)
		g.walk(stmt.Cond)
		g.writer.Cmp("0", RAX)
		g.writer.Je(lblEnd) // RAX???0(false)??????while?????????????????????
		g.walk(stmt.Body)
		g.writer.Jmp(lblBegin)
		g.writer.Label(lblEnd)
	case *ast.IfStmt:
		stmt, _ := node.(*ast.IfStmt)
		lblElse := g.genLbl()
		lblEnd := g.genLbl()
		g.walk(stmt.Cond)
		g.writer.Cmp("0", RAX)
		g.writer.Je(lblElse) // RAX???0(false)??????else???????????????????????????
		g.walk(stmt.IfBody)
		g.writer.Jmp(lblEnd)
		g.writer.Label(lblElse)
		if stmt.ElseBody != nil {
			g.walk(stmt.ElseBody)
		}
		g.writer.Label(lblEnd)
	case *ast.NumExp:
		val := fmt.Sprintf("%d", ty.Val)
		g.writer.Mov(val, getReg(RAX, ty.Type()))
	case *ast.IndexExp:
		g.address(g.currentFn, ty) // ???????????????????????????????????????????????????RAX?????????

		if ty.Type() == types.GetChar() {
			g.writer.Movsx("BYTE PTR "+g.writer.Address(RAX), EAX)
		} else {
			g.writer.Mov(g.writer.Address(RAX), getReg(RAX, ty.Type()))
		}
	case *ast.IdentExp:
		// ??????????????????
		g.address(g.currentFn, ty)

		if ty.Type() == types.GetChar() {
			g.writer.Movsx("BYTE PTR "+g.writer.Address(RAX), EAX)
		} else {
			g.writer.Mov(g.writer.Address(RAX), getReg(RAX, ty.Type()))
		}
	case *ast.StringLiteralExp:
		g.writer.Lea(ty.Label, RIP, RAX)
	case *ast.ArrayLiteral:
		for _, exp := range ty.AsInfixExps() {
			g.walk(exp)
		}
	case *ast.FuncCallExp:
		// ??????????????????
		defined := ty.Def != nil
		if !defined {
			// ????????????????????????????????????????????????????????????????????????????????????
		}

		for i, param := range ty.Params.Exps {
			if defined {
				arg := ty.Def.Args.LV.Locals[i]
				if !arg.Type.CanAssign(param.Type()) {
					g.Error(
						param.Token(),
						"Param types do not match for %s. Expected %s, but got %s.",
						arg.Name, arg.Type.String(), param.Type().String(),
					)
				}
			}

			g.walk(param)
			if i >= len(FUNCCALLREGS) {
				g.writer.Push(RAX)
			} else {
				g.writer.Mov(getReg(RAX, param.Type()), getReg(FUNCCALLREGS[i], param.Type()))
			}
		}

		// ???????????????????????????????????????????????????AL????????????????????????????????????????????????
		g.writer.Mov("0", AL)
		// FIXME: need align before call?
		g.writer.Call(ty.Name)
	case *ast.FuncDefNode:
		g.fns[ty.Name] = ty
		g.currentFn = ty
		g.writer.Globl(ty.Name)
		g.writer.Label(ty.Name)
		g.prolog()

		fn, ok := g.fns[ty.Name]
		if !ok {
			g.Error(ty.Token(), "Function %s not found.\n", ty.Name)
		}

		// Prepare params
		for i, local := range ty.Args.LV.Locals {
			offset, base := g.getOffset(fn, local)
			if i >= len(FUNCCALLREGS) {
				g.writer.Pop(RDI)
				g.writer.Lea(offset, RBP, RAX)
				g.writer.Mov(getReg(RDI, local.Type), g.writer.Address(RAX))
			} else {
				g.writer.Lea(offset, base, RAX)
				g.writer.Mov(getReg(FUNCCALLREGS[i], local.Type), g.writer.Address(RAX))
			}
		}

		g.walk(ty.Body)
		g.writer.Label(fmt.Sprintf(".L.return.%s", ty.Name))
		g.epilog()
	case *ast.UnaryExp:
		debug("Op:\t%s", ty.Op)
		switch ty.Op {
		case "&":
			g.address(g.currentFn, ty.Right) // RAX?????????????????????????????????
		case "*":
			g.walk(ty.Right) // RAX?????????????????????????????????
			g.writer.Mov(g.writer.Address(RAX), getReg(RAX, ty.Right.Type()))
		case "+":
			// do nothing ( +5 -> 5)
		case "-":
			g.walk(ty.Right)
			g.writer.Neg(RAX)
		case "sizeof":
			size := reduceSizeof(ty)
			g.writer.Mov(fmt.Sprintf("%d", size), EAX)
		}
	case *ast.DeclarationStmt:
		for _, local := range ty.LV.Locals {
			if ty.Exp != nil {
				// XXX: ????????????????????????
				// ???????????????????????????????????????????????????????????????RAX????????????
				switch ty.Exp.(type) {
				case *ast.ArrayLiteral:
					g.walk(ty.Exp)
				default:
					g.address(g.currentFn, local)
					g.writer.Push(RAX) // ??????2??????RAX???????????????????????????RAX????????????????????????
					g.walk(ty.Exp)
					g.writer.Pop(RDI)
					g.writer.Mov(getReg(RAX, local.Type), g.writer.Address(RDI))
				}
			}
		}
		// ????????????RAX??????????????????
	case *ast.InfixExp:
		debug("Op:\t %s", ty.Op)
		infix := ty

		if infix.Op == "=" {
			g.address(g.currentFn, infix.Left)
			g.writer.Push(RAX) // ??????2??????RAX???????????????????????????RAX????????????????????????
			g.walk(infix.Right)
			g.writer.Pop(RDI)
			g.writer.Mov(getReg(RAX, infix.Right.Type()), g.writer.Address(RDI))
			return
		}

		g.walk(infix.Right) // ????????????????????????RDI??????????????????????????????????????????-????????????
		g.writer.Push(RAX)
		g.walk(infix.Left)
		g.writer.Pop(RDI)

		switch infix.Op {
		case "+":
			switch leftTy := infix.Left.Type().(type) {
			case *types.Array:
				// RDI???right??????????????????????????????????????????
				unit := leftTy.Base.StackSize()
				g.writer.Mul(fmt.Sprint(unit), RDI)
				g.writer.Neg(RDI)
			case *types.IntPointer:
				// RDI???right??????????????????????????????????????????
				unit := leftTy.Base.StackSize()
				g.writer.Mul(fmt.Sprint(unit), RDI)
				g.writer.Neg(RDI)
			}
			g.writer.Add(RDI, RAX)
		case "-":
			switch leftTy := infix.Left.Type().(type) {
			case *types.Array:
				// RDI???right??????????????????????????????????????????
				unit := leftTy.Base.StackSize()
				g.writer.Mul(fmt.Sprint(unit), RDI)
				g.writer.Neg(RDI)
			case *types.IntPointer:
				// RDI???right??????????????????????????????????????????
				unit := leftTy.Base.StackSize()
				g.writer.Mul(fmt.Sprint(unit), RDI)
				g.writer.Neg(RDI)
			}
			g.writer.Sub(RDI, RAX) // ?????????RDI????????????????????????
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

func (g *Generator) getOffset(fn *ast.FuncDefNode, node interface{}) (string, string) {
	var name string
	var ty types.Type

	switch node := node.(type) {
	case *ast.LocalVariable:
		name = node.Name
		ty = node.Type
	case *ast.IdentExp:
		name = node.Name
		ty = node.Type()
	default:
		err("Invalid node: %T\n", node)
		os.Exit(1)
	}

	// local
	offset, ok := fn.Offsets[name]
	if ok {
		return fmt.Sprintf("-%d", offset), RBP
	}

	// global
	_, ok = g.globals()[name]
	if ok {
		return name, RIP
	}

	err("Invalid ident name: '%s' of type '%s'\n", name, ty)
	os.Exit(1)
	return "", RBP
}

// FIXME: Dereference is not supported. i.e. int *x = &y
func (g *Generator) eval(exp ast.Exp) ast.Exp {
	debug("eval %T, %s", exp, exp)
	switch exp := exp.(type) {
	case *ast.NumExp:
		return exp
	case *ast.StringLiteralExp:
		return exp
	case *ast.ArrayLiteral:
		return exp
	case *ast.UnaryExp:
		if exp.Op == "&" {
			_, ok := exp.Right.(*ast.IdentExp)
			if !ok {
				g.Error(exp.Right.Token(), "Invalid exp for rvalue of & unary exp: %s", exp.Right)
			}
			return exp
		}

		right := g.eval(exp.Right)
		r, ok := right.(*ast.NumExp)
		if !ok {
			g.Error(right.Token(), "Invalid right for unary exp: %s", right)
		}

		if exp.Op == "+" {
			return r
		}

		if exp.Op == "-" {
			return &ast.NumExp{Val: -r.Val}
		}

		g.Error(exp.Token(), "Invalid operator for global rvalue unary right: %s", exp.Op)
	case *ast.InfixExp:
		debug("eval %T, %s", exp, exp)
		left := g.eval(exp.Left)
		right := g.eval(exp.Right)
		l, ok := left.(*ast.NumExp)
		if !ok {
			g.Error(left.Token(), "Invalid left exp: %s", left)
		}

		r, ok := right.(*ast.NumExp)
		if !ok {
			g.Error(right.Token(), "Invalid right exp: %s", right)
		}

		if exp.Op == "+" {
			val := l.Val + r.Val
			return &ast.NumExp{Val: val}
		}

		if exp.Op == "-" {
			val := l.Val - r.Val
			return &ast.NumExp{Val: val}
		}

		if exp.Op == "*" {
			val := l.Val * r.Val
			return &ast.NumExp{Val: val}
		}

		if exp.Op == "/" {
			val := l.Val / r.Val
			return &ast.NumExp{Val: val}
		}

		g.Error(exp.Token(), "Invalid operator for global rvalue: %s", exp.Op)
	}

	g.Error(exp.Token(), "Invalid exp for global rvalue: %s", exp)
	return nil
}

func reduceSizeof(unary *ast.UnaryExp) int {
	return unary.Right.Type().StackSize()
}

func debug(s string, args ...interface{}) {
	if DEBUG {
		err(s, args...)
	}
}

func err(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s+"\n", args...)
}

func getReg(reg string, ty types.Type) string {
	switch ty.(type) {
	case *types.Char:
		r, ok := BYTE[reg]
		if !ok {
			goto ERROR
		}
		return r
	case *types.Int:
		r, ok := DWORD[reg]
		if !ok {
			goto ERROR
		}
		return r
	case *types.Array:
		return reg
	case *types.IntPointer:
		return reg
	default:
		goto ERROR
	}

ERROR:
	err("Invalid size %T for %s", ty, reg)
	os.Exit(1)
	return ""
}

func getType(size int) string {
	switch size {
	case 1:
		return ".byte"
	case 4:
		return ".long"
	case 8:
		return ".quad"
	}

	err("Invalid data size: %d", size)
	return ""
}

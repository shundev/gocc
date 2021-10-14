package types

import (
	"bytes"
	"fmt"
)

const (
	INT         = "INT"
	INT_POINTER = "INT_POINTER"
)

var (
	int_  = &Int{}
	char_ = &Char{}
)

type Type interface {
	String() string
	Size() int      // それを指す参照のサイズ
	StackSize() int // データが実際にメモリを占めるサイズ
	CanAssign(right Type) bool
	CanAdd(right Type) bool
	CanMul(right Type) bool
}

type Char struct {
}

func (t *Char) String() string {
	return "char"
}

func (t *Char) Size() int {
	return 1
}

func (t *Char) StackSize() int {
	return 1
}

func (t *Char) CanAssign(right Type) bool {
	return right == int_ || right == char_
}

func (t *Char) CanAdd(right Type) bool {
	return right == int_ || right == char_
}

func (t *Char) CanMul(right Type) bool {
	return right == int_ || right == char_
}

type Int struct {
}

func (t *Int) String() string {
	return "int"
}

func (t *Int) Size() int {
	return 4
}

func (t *Int) StackSize() int {
	return 4
}

func (t *Int) CanAssign(right Type) bool {
	return right == int_ || right == char_
}

func (t *Int) CanAdd(right Type) bool {
	return right == int_ || right == char_
}

func (t *Int) CanMul(right Type) bool {
	return right == int_ || right == char_
}

type IntPointer struct {
	Base Type
}

func (t *IntPointer) String() string {
	s := "*"
	base := t.Base
	for {
		switch ty := base.(type) {
		case *Int:
			s = "int" + s
			return s
		case *IntPointer:
			s = "*" + s
			base = ty.Base
		}
	}
}

func (t *IntPointer) Size() int {
	return 8
}

func (t *IntPointer) StackSize() int {
	return 8
}

func (t *IntPointer) CanAssign(right Type) bool {
	_, ok := right.(*IntPointer)
	if ok {
		return true
	}

	_, ok = right.(*Array)
	return ok
}

func (t *IntPointer) CanAdd(right Type) bool {
	return right == int_ || right == char_
}

func (t *IntPointer) CanMul(right Type) bool {
	return false
}

type Array struct {
	Base   Type
	Length int
}

func (t *Array) String() string {
	var out bytes.Buffer
	out.WriteString(t.Base.String())
	out.WriteString(fmt.Sprintf("[%d]", t.Length))
	return out.String()
}

func (t *Array) Size() int {
	return 8
}

func (t *Array) StackSize() int {
	return t.Base.StackSize() * t.Length
}

func (t *Array) CanAssign(right Type) bool {
	arr, ok := right.(*Array)
	if !ok {
		return false
	}

	if arr.Length != t.Length {
		return false
	}

	if arr.Base != t.Base {
		return false
	}

	return true
}

func (t *Array) CanAdd(right Type) bool {
	return right == int_ || right == char_
}

func (t *Array) CanMul(right Type) bool {
	return false
}

/* Factory */

func GetInt() Type {
	return int_
}

func GetChar() Type {
	return char_
}

func PointerTo(base Type) Type {
	return &IntPointer{Base: base}
}

func ArrayOf(base Type, length int) Type {
	return &Array{Base: base, Length: length}
}

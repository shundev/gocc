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
	int_ = &Int{}
)

type Type interface {
	String() string
	Size() int
	StackSize() int
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
	return 8
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
	return t.Size()
}

func GetInt() Type {
	return int_
}

func PointerTo(base Type) Type {
	return &IntPointer{Base: base}
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
	return t.Base.Size() * t.Length
}

func (t *Array) StackSize() int {
	return t.Base.StackSize() * t.Length
}

func ArrayOf(base Type, length int) Type {
	return &Array{Base: base, Length: length}
}

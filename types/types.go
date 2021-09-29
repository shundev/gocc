package types

import "fmt"

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
}

type Int struct {
}

func (t *Int) String() string {
	return "int"
}

func (t *Int) Size() int {
	return 4
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
	base := t.Base
	s := ""
	loop := true
	for loop {
		switch ty := base.(type) {
		case *Int:
			s = "int" + s
			loop = false
		case *IntPointer:
			s = "*" + s
			base = ty.Base
		}
	}

	s += fmt.Sprintf("[%d]", t.Length)
	return s
}

func (t *Array) Size() int {
	return t.Base.Size() * t.Length
}

func ArrayOf(base Type, length int) Type {
	return &Array{Base: base, Length: length}
}

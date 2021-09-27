package types

const (
	INT         = "INT"
	INT_POINTER = "INT_POINTER"
)

var (
	int_ = &Int{}
)

type Type interface {
	String() string
}

type Int struct {
}

func (t *Int) String() string {
	return "int"
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

func GetInt() Type {
	return int_
}

func PointerTo(base Type) Type {
	return &IntPointer{Base: base}
}

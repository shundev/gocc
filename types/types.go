package types

const (
	INT         = "int"
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
	return INT
}

type IntPointer struct {
	Base Type
}

func (t *IntPointer) String() string {
	return INT_POINTER
}

func GetInt() Type {
	return int_
}

func PointerTo(base Type) Type {
	return &IntPointer{Base: base}
}

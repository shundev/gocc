package types

const (
	INT         = "INT"
	INT_POINTER = "INT_POINTER"
)

var (
	Int        = &int_{}
	IntPointer = &intPointer{}
)

type Type interface {
	String() string
}

type int_ struct{}

func (t *int_) String() string {
	return INT
}

type intPointer struct{}

func (t *intPointer) String() string {
	return INT_POINTER
}

package lib

import "fmt"

type V int

type LibStruct struct {
	V V
}

func (v LibStruct) Print() {
	fmt.Println(v.V)
}

var LibStruct1 = LibStruct{}
var Foo1 = 0
var Foo2 = 0

func init() {
	Foo1 = 10
}

const (
	HOGE1 = 1
	HOGE2 = 1
)

type Writer interface {
	Write()
}

type Reader interface {
	Reade()
}

func LibFunc() {
	fmt.Println("from lib")
}

type Seeker[T any] struct {
}

func NewSeeker[T any]() Seeker[T] {
	return Seeker[T]{}
}

func (s Seeker[T]) Seek() {
	fmt.Println("seeker is seeking")
}

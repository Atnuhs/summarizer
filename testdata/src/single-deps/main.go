package main

import (
	"fmt"

	"github.com/Atnuhs/go-bundler/testdata/src/single-deps/lib"
)

func init_sub() {

}

func init() {
	fmt.Println("hoge")
	init_sub()
}

const (
	X1 = iota
	X2
	X3
)

var (
	Foo1 = lib.Foo1
	Foo2 = lib.Foo2
)

const HOGE11, HOGE12 = lib.HOGE1, lib.HOGE2
const HOGE2 = lib.HOGE2

// Embedded is Embeddding another package struct
type Embedded struct {
	lib.LibStruct
}

func (d Embedded) String() {
	fmt.Println(d.LibStruct.V)
}

type NonEmbedded struct {
	s lib.LibStruct
}

func (d NonEmbedded) String() {
	fmt.Println(d.s.V)
}

type Seeker interface {
	Seek()
}

type RWSer interface {
	lib.Reader
	lib.Writer
	Seeker
}

func SeekerSeek(s Seeker) {
	s.Seek()
}

func FunctionWithArg(x int) {
	var inner = 1
	fmt.Println(x, inner)
}

func main() {
	lib.LibFunc()
	lib.LibStruct1.V = 10
	data := Embedded{lib.LibStruct{}}
	data2 := Embedded{LibStruct: lib.LibStruct{}}
	data3 := NonEmbedded{s: lib.LibStruct{}}
	fmt.Println(data.V)
	fmt.Println(data2.V)
	fmt.Println(data3.s.V)
	fmt.Println(HOGE11)
	fmt.Println(X1)
	FunctionWithArg(10)
	SeekerSeek(lib.NewSeeker[int]())

}

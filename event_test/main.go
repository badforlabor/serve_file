/**
 * Auth :   liubo
 * Date :   2019/11/11 11:52
 * Comment:
 */

package main

import (
	"fmt"
	"serve_file/proto"
	"unsafe"
)

type classA struct {
	v string
}
func (self *classA) show() {
	fmt.Println(self.v)
}

type Pointer2 struct {
	v1 unsafe.Pointer
	v2 unsafe.Pointer
}

type strFunc *func()

func show1() {
	fmt.Println("show1")
}
func show2() {
	fmt.Println("show2")
}

func main()  {
	var a classA
	var b classA

	a.v = "a"
	b.v = "b"

	var e proto.FunctionEvent
	e.Add(a, a.show)
	e.Add(b, b.show)
	e.Add(nil, show1)
	e.Add(nil, show2)
	e.Broadcast()

	fmt.Println("--------- 1 -----------")
	e.Remove(b, b.show)
	e.Broadcast()

	fmt.Println("--------- 2 -----------")
	e.Remove(b, b.show)
	e.Broadcast()

	fmt.Println("--------- 3 -----------")
	e.Remove(a, a.show)
	e.Broadcast()

	fmt.Println("--------- 4 -----------")
	e.Remove(nil, show1)
	e.Broadcast()

	fmt.Println("--------- 5 -----------")
	e.Remove(nil, show2)
	e.Broadcast()
}

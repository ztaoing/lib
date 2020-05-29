package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func main() {
	str := "abc"
	fmt.Printf("%08x", stringAddr(str))
}
func stringAddr(s string) uintptr {
	return (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
}

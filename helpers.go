package gruby

import "unsafe"

// #include "gruby.h"
import "C"

func freeStr(cstr *C.char) {
	C.free(unsafe.Pointer(cstr))
}

func Must0[T any](err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

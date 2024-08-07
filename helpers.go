package gruby

import "unsafe"

// #include "gruby.h"
import "C"

func freeStr(cstr *C.char) {
	C.free(unsafe.Pointer(cstr))
}

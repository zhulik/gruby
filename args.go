package gruby

// #include "gruby.h"
import "C"

// ArgSpec defines how many arguments a function should take and
// what kind. Multiple ArgSpecs can be combined using the "|"
// operator.
type ArgSpec C.mrb_aspec

// ArgsAny allows any number of arguments.
func ArgsAny() ArgSpec {
	return ArgSpec(C._go_MRB_ARGS_ANY())
}

// ArgsArg says the given number of arguments are required and
// the second number is optional.
func ArgsArg(r, o int) ArgSpec {
	return ArgSpec(C._go_MRB_ARGS_ARG(C.int(r), C.int(o)))
}

// ArgsBlock says it takes a block argument.
func ArgsBlock() ArgSpec {
	return ArgSpec(C._go_MRB_ARGS_BLOCK())
}

// ArgsNone says it takes no arguments.
func ArgsNone() ArgSpec {
	return ArgSpec(C._go_MRB_ARGS_NONE())
}

// ArgsReq says that the given number of arguments are required.
func ArgsReq(n int) ArgSpec {
	return ArgSpec(C._go_MRB_ARGS_REQ(C.int(n)))
}

// ArgsOpt says that the given number of arguments are optional.
func ArgsOpt(n int) ArgSpec {
	return ArgSpec(C._go_MRB_ARGS_OPT(C.int(n)))
}

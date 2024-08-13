package gruby

// #include "gruby.h"
import "C"

import (
	"strconv"
	"strings"
	"unsafe"
)

// ExceptionError is a special type of value that represents an error
// and implements the Error interface.
type ExceptionError struct {
	Value
	File      string
	Line      int
	Message   string
	Backtrace []string
}

func (e *ExceptionError) Error() string {
	return e.Message
}

//-------------------------------------------------------------------
// Internal Functions
//-------------------------------------------------------------------

func newExceptionValue(grb *GRuby) *ExceptionError {
	state := grb.state
	if state.exc == nil {
		return nil
	}

	arenaIndex := C._go_mrb_gc_arena_save(state)
	defer C._go_mrb_gc_arena_restore(state, arenaIndex)

	// Convert the RObject* to an mrb_value
	value := C.mrb_obj_value(unsafe.Pointer(state.exc))

	// Retrieve and convert backtrace to []string (avoiding reflection in Decode)
	var backtrace []string
	grbBacktraceValue := grb.value(C.mrb_exc_backtrace(state, value))
	if grbBacktraceValue.Type() == TypeArray {
		grbBacktrace := MustToGo[Values](grbBacktraceValue)
		for _, ln := range grbBacktrace {
			backtrace = append(backtrace, ln.String())
		}
	}

	// Extract file + line from first backtrace line
	file := "Unknown"
	line := 0
	if len(backtrace) > 0 {
		fileAndLine := strings.Split(backtrace[0], ":")
		if len(fileAndLine) >= 2 { //nolint:mnd
			file = fileAndLine[0]
			line = Must(strconv.Atoi(fileAndLine[1]))
		}
	}

	result := grb.value(value)
	return &ExceptionError{
		Value:     result,
		Message:   result.String(),
		File:      file,
		Line:      line,
		Backtrace: backtrace,
	}
}

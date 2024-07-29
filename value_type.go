package mruby

// #include <stdlib.h>
// #include "gomruby.h"
import "C"

// ValueType is an enum of types that a Value can be and is returned by
// Value.Type().
type ValueType uint32

const (
	// TypeFalse is `false`
	TypeFalse = C.MRB_TT_FALSE
	// TypeTrue is `true`
	TypeTrue = C.MRB_TT_TRUE
	// TypeFloat is any floating point number such as 1.2, etc.
	TypeFloat = C.MRB_TT_FLOAT
	// TypeFixnum is fixnums, or integers for this case.
	TypeFixnum = C.MRB_TT_FIXNUM
	// TypeSymbol is for entities in ruby that look like `:this`
	TypeSymbol = C.MRB_TT_SYMBOL
	// TypeUndef is a value internal to ruby for uninstantiated vars.
	TypeUndef = C.MRB_TT_UNDEF
	// TypeCptr is a void*
	TypeCptr = C.MRB_TT_CPTR
	// TypeFree is ?
	TypeFree = C.MRB_TT_FREE
	// TypeClass is the base class of all classes.
	TypeObject = C.MRB_TT_OBJECT
	// TypeClass is the base class of all classes.
	TypeClass = C.MRB_TT_CLASS
	// TypeModule is the base class of all Modules.
	TypeModule = C.MRB_TT_MODULE
	// TypeIClass is ?
	TypeIClass = C.MRB_TT_ICLASS
	// TypeSClass is ?
	TypeSClass = C.MRB_TT_SCLASS
	// TypeProc are procs (concrete block definitons)
	TypeProc = C.MRB_TT_PROC
	// TypeArray is []
	TypeArray = C.MRB_TT_ARRAY
	// TypeHash is { }
	TypeHash = C.MRB_TT_HASH
	// TypeString is ""
	TypeString = C.MRB_TT_STRING
	// TypeRange is (0..x)
	TypeRange = C.MRB_TT_RANGE
	// TypeException is raised when using the raise keyword
	TypeException = C.MRB_TT_EXCEPTION
	// TypeEnv is for getenv/setenv etc
	TypeEnv = C.MRB_TT_ENV
	// TypeData is ?
	TypeData = C.MRB_TT_DATA
	// TypeFiber is for members of the Fiber class
	TypeFiber = C.MRB_TT_FIBER
	// TypeIsStruct is ?
	TypeIsStruct = C.MRB_TT_STRUCT
	// TypeMaxBreak is ?
	TypeBreak = C.MRB_TT_BREAK
	// TypeMaxDefine is ?
	TypeMaxDefine = C.MRB_TT_MAXDEFINE
	// TypeNil is nil
	TypeNil ValueType = 0xffffffff
)

package gruby

// #include "gruby.h"
import "C"

// ValueType is an enum of types that a Value can be and is returned by
// Value.Type().
type ValueType uint32

const (
	// TypeFalse is `false`
	TypeFalse = ValueType(C.MRB_TT_FALSE)
	// TypeTrue is `true`
	TypeTrue = ValueType(C.MRB_TT_TRUE)
	// TypeFloat is any floating point number such as 1.2, etc.
	TypeFloat = ValueType(C.MRB_TT_FLOAT)
	// TypeFixnum is fixnums, or integers for this case.
	TypeFixnum = ValueType(C.MRB_TT_FIXNUM)
	// TypeSymbol is for entities in ruby that look like `:this`
	TypeSymbol = ValueType(C.MRB_TT_SYMBOL)
	// TypeUndef is a value internal to ruby for uninstantiated vars.
	TypeUndef = ValueType(C.MRB_TT_UNDEF)
	// TypeCptr is a void*
	TypeCptr = ValueType(C.MRB_TT_CPTR)
	// TypeFree is ?
	TypeFree = ValueType(C.MRB_TT_FREE)
	// TypeClass is the base class of all classes.
	TypeObject = ValueType(C.MRB_TT_OBJECT)
	// TypeClass is the base class of all classes.
	TypeClass = ValueType(C.MRB_TT_CLASS)
	// TypeModule is the base class of all Modules.
	TypeModule = ValueType(C.MRB_TT_MODULE)
	// TypeIClass is ?
	TypeIClass = ValueType(C.MRB_TT_ICLASS)
	// TypeSClass is ?
	TypeSClass = ValueType(C.MRB_TT_SCLASS)
	// TypeProc are procs (concrete block definitons)
	TypeProc = ValueType(C.MRB_TT_PROC)
	// TypeArray is []
	TypeArray = ValueType(C.MRB_TT_ARRAY)
	// TypeHash is { }
	TypeHash = ValueType(C.MRB_TT_HASH)
	// TypeString is ""
	TypeString = ValueType(C.MRB_TT_STRING)
	// TypeRange is (0..x)
	TypeRange = ValueType(C.MRB_TT_RANGE)
	// TypeException is raised when using the raise keyword
	TypeException = ValueType(C.MRB_TT_EXCEPTION)
	// TypeEnv is for getenv/setenv etc
	TypeEnv = ValueType(C.MRB_TT_ENV)
	// TypeData is ?
	TypeData = ValueType(C.MRB_TT_DATA)
	// TypeFiber is for members of the Fiber class
	TypeFiber = ValueType(C.MRB_TT_FIBER)
	// TypeIsStruct is ?
	TypeIsStruct = ValueType(C.MRB_TT_STRUCT)
	// TypeMaxBreak is ?
	TypeBreak = ValueType(C.MRB_TT_BREAK)
	// TypeMaxDefine is ?
	TypeMaxDefine = ValueType(C.MRB_TT_MAXDEFINE)
	// TypeNil is nil
	TypeNil ValueType = 0xffffffff
)

package gruby

// #cgo CFLAGS: -Imruby-build/mruby/include
// #cgo LDFLAGS: ${SRCDIR}/libmruby.a -lm
// #include <stdlib.h>
// #include "gruby.h"
import "C"

import (
	"strings"
	"unsafe"
)

type (
	classMethodMap map[*C.struct_RClass]methodMap
	methodMap      map[C.mrb_sym]Func
)

// GRuby represents a single instance of gruby.
type GRuby struct {
	state *C.mrb_state

	loadedFiles       map[string]bool
	classes           classMethodMap
	getArgAccumulator []C.mrb_value
}

//export goGetArgAppend
func goGetArgAppend(state *C.mrb_state, v C.mrb_value) {
	grb := states.Get(state)

	grb.getArgAccumulator = append(grb.getArgAccumulator, v)
}

// GetGlobalVariable returns the value of the global variable by the given name.
func (g *GRuby) GetGlobalVariable(name string) Value {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	return g.value(C._go_mrb_gv_get(g.state, C.mrb_intern_cstr(g.state, cstr)))
}

// SetGlobalVariable sets the value of the global variable by the given name.
func (g *GRuby) SetGlobalVariable(name string, value Value) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	C._go_mrb_gv_set(g.state, C.mrb_intern_cstr(g.state, cstr), value.CValue())
}

// ArenaIndex represents the index into the arena portion of the GC.
//
// See ArenaSave for more information.
type ArenaIndex int

// New creates a new instance of Mrb, representing the state of a single
// Ruby VM.
//
// When you're finished with the VM, clean up all resources it is using
// by calling the Close method.
func New() *GRuby {
	state := C.mrb_open()

	grb := &GRuby{
		state:             state,
		loadedFiles:       map[string]bool{},
		classes:           classMethodMap{},
		getArgAccumulator: make([]C.mrb_value, 0, C._go_get_max_funcall_args()),
	}

	states.Add(state, grb)

	return grb
}

// ArenaRestore restores the arena index so the objects between the save and this point
// can be garbage collected in the future.
//
// See ArenaSave for more documentation.
func (g *GRuby) ArenaRestore(idx ArenaIndex) {
	C._go_mrb_gc_arena_restore(g.state, C.int(idx))
}

// ArenaSave saves the index into the arena.
//
// Restore the arena index later by calling ArenaRestore.
//
// The arena is where objects returned by functions such as LoadString
// are stored. By saving the index and then later restoring it with
// ArenaRestore, these objects can be garbage collected. Otherwise, the
// objects will never be garbage collected.
//
// The recommended usage pattern for memory management is to save
// the arena index prior to any Ruby execution, to turn the resulting
// Ruby value into Go values as you see fit, then to restore the arena
// index so that GC can collect any values.
//
// Of course, when Close() is called, all objects in the arena are
// garbage collected anyways, so if you're only calling mruby for a short
// period of time, you might not have to worry about saving/restoring the
// arena.
func (g *GRuby) ArenaSave() ArenaIndex {
	return ArenaIndex(C._go_mrb_gc_arena_save(g.state))
}

// EnableGC enables the garbage collector for this mruby instance. It returns
// true if garbage collection was previously disabled.
func (g *GRuby) EnableGC() {
	C._go_enable_gc(g.state)
}

// DisableGC disables the garbage collector for this mruby instance. It returns
// true if it was previously disabled.
func (g *GRuby) DisableGC() {
	C._go_disable_gc(g.state)
}

// LiveObjectCount returns the number of objects that have not been collected (aka, alive).
func (g *GRuby) LiveObjectCount() int {
	return int(C._go_gc_live(g.state))
}

// Class returns the class with the kgiven name and superclass. Note that
// if you call this with a class that doesn't exist, mruby will abort the
// application (like a panic, but not a Go panic).
//
// super can be nil, in which case the Object class will be used.
func (g *GRuby) Class(name string, super *Class) *Class {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	var class *C.struct_RClass
	if super == nil {
		class = C.mrb_class_get(g.state, cstr)
	} else {
		class = C.mrb_class_get_under(g.state, super.class, cstr)
	}

	return newClass(g, class)
}

// Module returns the named module as a *Class. If the module is invalid,
// NameError is triggered within your program and SIGABRT is sent to the
// application.
func (g *GRuby) Module(name string) *Class {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	class := C.mrb_module_get(g.state, cstr)

	return newClass(g, class)
}

// Close a Mrb, this must be called to properly free resources, and
// should only be called once.
func (g *GRuby) Close() {
	states.Delete(g.state)
	C.mrb_close(g.state)
}

// ConstDefined checks if the given constant is defined in the scope.
//
// This should be used, for example, before a call to Class, because a
// failure in Class will crash your program (by design). You can retrieve
// the Value of a Class by calling Value().
func (g *GRuby) ConstDefined(name string, scope Value) bool {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	scopeV := scope.CValue()
	b := C.mrb_const_defined(
		g.state, scopeV, C.mrb_intern_cstr(g.state, cstr))

	// TODO: a go helper function?
	return C._go_mrb_bool2int(b) != 0
}

// FullGC executes a complete GC cycle on the VM.
func (g *GRuby) FullGC() {
	C.mrb_full_gc(g.state)
}

// GetArgs returns all the arguments that were given to the currnetly
// called function (currently on the stack).
func (g *GRuby) GetArgs() []Value {
	// Clear reset the accumulator to zero length
	g.getArgAccumulator = make([]C.mrb_value, 0, C._go_get_max_funcall_args())

	// Get all the arguments and put it into our accumulator
	count := C._go_mrb_get_args_all(g.state)

	// Convert those all to values
	values := make([]Value, count)

	for i := range int(count) {
		values[i] = g.value(g.getArgAccumulator[i])
	}

	return values
}

// IncrementalGC runs an incremental GC step. It is much less expensive
// than a FullGC, but must be called multiple times for GC to actually
// happen.
//
// This function is best called periodically when executing Ruby in
// the VM many times (thousands of times).
func (g *GRuby) IncrementalGC() {
	C.mrb_incremental_gc(g.state)
}

// LoadString loads the given code, executes it, and returns its final
// value that it might return.
func (g *GRuby) LoadString(code string) (Value, error) {
	cstr := C.CString(code)
	defer C.free(unsafe.Pointer(cstr))

	value := C._go_mrb_load_string(g.state, cstr)
	if exc := checkException(g); exc != nil {
		return nil, exc
	}

	return g.value(value), nil
}

// LoadStringWith loads the given code, executes it within the given context, and returns its final
// value that it might return.
func (g *GRuby) LoadStringWithContext(code string, ctx *CompileContext) (Value, error) {
	cstr := C.CString(code)
	defer C.free(unsafe.Pointer(cstr))

	// TODO:
	/** program load functions
	* Please note! Currently due to interactions with the GC calling these functions will
	* leak one RProc object per function call.
	* To prevent this save the current memory arena before calling and restore the arena
	* right after, like so
	* int ai = mrb_gc_arena_save(mrb);
	* mrb_value status = mrb_load_string(mrb, buffer);
	* mrb_gc_arena_restore(mrb, ai);
	 */

	value := C.mrb_load_string_cxt(g.state, cstr, ctx.ctx)
	if exc := checkException(g); exc != nil {
		return nil, exc
	}

	return g.value(value), nil
}

// Run executes the given value, which should be a proc type.
//
// If you're looking to execute code directly a string, look at LoadString.
//
// If self is nil, it is set to the top-level self.
func (g *GRuby) Run(v Value, self Value) (Value, error) {
	if self == nil {
		self = g.TopSelf()
	}

	mrbV := v.CValue()
	mrbSelf := self.CValue()

	proc := C._go_mrb_proc_ptr(mrbV)
	value := C.mrb_vm_run(g.state, proc, mrbSelf, 0)

	if exc := checkException(g); exc != nil {
		return nil, exc
	}

	return g.value(value), nil
}

// RunWithContext is a context-aware parser (aka, it does not discard state
// between runs). It returns a magic integer that describes the stack in place,
// so that it can be re-used on the next call. This is how local variables can
// traverse ruby parse invocations.
//
// Otherwise, it is very similar in function to Run()
func (g *GRuby) RunWithContext(v Value, self Value, stackKeep int) (int, Value, error) {
	if self == nil {
		self = g.TopSelf()
	}

	mrbV := v.CValue()
	mrbSelf := self.CValue()
	proc := C._go_mrb_proc_ptr(mrbV)

	keep := C.int(stackKeep)

	value := C._go_mrb_vm_run(g.state, proc, mrbSelf, &keep)

	if exc := checkException(g); exc != nil {
		return stackKeep, nil, exc
	}

	return int(keep), g.value(value), nil
}

// Yield yields to a block with the given arguments.
//
// This should be called within the context of a Func.
func (g *GRuby) Yield(block Value, args ...Value) (Value, error) {
	mrbBlock := block.CValue()

	var argv []C.mrb_value
	var argvPtr *C.mrb_value

	if len(args) > 0 {
		// Make the raw byte slice to hold our arguments we'll pass to C
		argv = make([]C.mrb_value, len(args))
		for i, arg := range args {
			argv[i] = arg.CValue()
		}

		argvPtr = &argv[0]
	}

	result := C._go_mrb_yield_argv(
		g.state,
		mrbBlock,
		C.mrb_int(len(argv)),
		argvPtr)

	if exc := checkException(g); exc != nil {
		return nil, exc
	}

	return g.value(result), nil
}

//-------------------------------------------------------------------
// Functions handling defining new classes/modules in the VM
//-------------------------------------------------------------------

// DefineClass defines a new top-level class.
//
// If super is nil, the class will be defined under Object.
func (g *GRuby) DefineClass(name string, super *Class) *Class {
	if super == nil {
		super = g.ObjectClass()
	}

	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	return newClass(g, C.mrb_define_class(g.state, cstr, super.class))
}

// DefineClassUnder defines a new class under another class.
//
// This is, for example, how you would define the World class in
// `Hello::World` where Hello is the "outer" class.
func (g *GRuby) DefineClassUnder(name string, super *Class, outer *Class) *Class {
	if super == nil {
		super = g.ObjectClass()
	}
	if outer == nil {
		outer = g.ObjectClass()
	}

	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	return newClass(g, C.mrb_define_class_under(
		g.state, outer.class, cstr, super.class))
}

// DefineModule defines a top-level module.
func (g *GRuby) DefineModule(name string) *Class {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	return newClass(g, C.mrb_define_module(g.state, cstr))
}

// DefineModuleUnder defines a module under another class/module.
func (g *GRuby) DefineModuleUnder(name string, outer *Class) *Class {
	if outer == nil {
		outer = g.ObjectClass()
	}

	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	return newClass(g,
		C.mrb_define_module_under(g.state, outer.class, cstr))
}

//-------------------------------------------------------------------
// Functions below return Values or constant Classes
//-------------------------------------------------------------------

// ObjectClass returns the Object top-level class.
func (g *GRuby) ObjectClass() *Class {
	return newClass(g, g.state.object_class)
}

// KernelModule returns the Kernel top-level module.
func (g *GRuby) KernelModule() *Class {
	return newClass(g, g.state.kernel_module)
}

// TopSelf returns the top-level `self` value.
func (g *GRuby) TopSelf() Value {
	return g.value(C.mrb_obj_value(unsafe.Pointer(g.state.top_self)))
}

// FalseValue returns a Value for "false"
func (g *GRuby) FalseValue() Value {
	return g.value(C.mrb_false_value())
}

// NilValue returns "nil"
func (g *GRuby) NilValue() Value {
	return g.value(C.mrb_nil_value())
}

// TrueValue returns a Value for "true"
func (g *GRuby) TrueValue() Value {
	return g.value(C.mrb_true_value())
}

// When called from a methods defined in Go, returns current ruby backtrace.
func (g *GRuby) Backtrace() []string {
	backtrace := g.value(C.mrb_get_backtrace(g.state))
	array := ToGo[*Array](backtrace)

	result := make([]string, array.Len())

	for i := range array.Len() {
		item, err := array.Get(i)
		if err != nil {
			panic(err)
		}
		result[i] = ToGo[string](item)
	}

	return result
}

// When called from a method defined in Go, returns a full name of a file the method was called from.
// Currently implemented using the backtrace.
// TODO: a better way?
func (g *GRuby) CalledFromFile() string {
	return strings.Split(g.Backtrace()[0], ":")[0]
}

func (g *GRuby) LoadFile(path string, content string) (bool, *CompileContext, error) {
	if g.loadedFiles[path] {
		return false, nil, nil
	}

	ctx := NewCompileContext(g)
	ctx.SetFilename(path)

	_, err := g.LoadStringWithContext(string(content), ctx)
	if err != nil {
		return false, ctx, err
	}

	g.loadedFiles[path] = true

	return true, ctx, nil
}

func (g *GRuby) value(v C.mrb_value) Value {
	return &MrbValue{
		grb:   g,
		value: v,
	}
}

func (g *GRuby) insertMethod(class *C.struct_RClass, name string, callback Func) {
	methodLookup := g.classes[class]
	if methodLookup == nil {
		methodLookup = make(methodMap)
		g.classes[class] = methodLookup
	}

	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	sym := C.mrb_intern_cstr(g.state, cstr)
	methodLookup[sym] = callback
}

func checkException(grb *GRuby) error {
	if grb.state.exc == nil {
		return nil
	}

	err := newExceptionValue(grb)
	grb.state.exc = nil

	return err
}

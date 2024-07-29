package mruby_test

import (
	"reflect"
	"testing"

	mruby "github.com/zhulik/gruby"
)

func TestExceptionString_afterClose(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	_, err := mrb.LoadString(`clearly a syntax error`)
	mrb.Close()
	// This panics before the bug fix that this test tests
	str := err.Error()
	if str != "undefined method 'error'" {
		t.Fatalf("'%s'", str)
	}
}

func TestExceptionBacktrace(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	parser := mruby.NewParser(mrb)
	defer parser.Close()
	context := mruby.NewCompileContext(mrb)
	context.SetFilename("hello.rb")
	defer context.Close()

	_, err := parser.Parse(`
				def do_error
					raise "Exception"
				end

				def hop1
					do_error
				end

				def hop2
					hop1
				end

				hop2
			`, context)
	if err != nil {
		t.Fatal(err)
	}

	proc := parser.GenerateCode()
	_, err = mrb.Run(proc, nil)
	if err == nil {
		t.Fatalf("expected exception")
	}

	exc := err.(*mruby.ExceptionError)
	if exc.Message != "Exception" {
		t.Fatalf("bad exception message: %s", exc.Message)
	}

	if exc.File != "hello.rb" {
		t.Fatalf("bad file: %s", exc.File)
	}

	if exc.Line != 3 {
		t.Fatalf("bad line: %d", exc.Line)
	}

	if len(exc.Backtrace) != 4 {
		t.Fatalf("bad backtrace: %#v", exc.Backtrace)
	}
}

func TestMrbValueCall(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"foo"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, err = value.Call("some_function_that_doesnt_exist")
	if err == nil {
		t.Fatalf("expected exception")
	}

	result, err := value.Call("==", mruby.ToRuby(mrb, "foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.Type() != mruby.TypeTrue {
		t.Fatalf("bad type")
	}
}

func TestMrbValueCallBlock(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"foo"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	block, err := mrb.LoadString(`Proc.new { |_| "bar" }`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := value.CallBlock("gsub", mruby.ToRuby(mrb, "foo"), block)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.Type() != mruby.TypeString {
		t.Fatalf("bad type")
	}
	if result.String() != "bar" {
		t.Fatalf("bad: %s", result)
	}
}

func TestMrbValueValue_impl(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	var _ mruby.Value = mrb.FalseValue()
}

func TestMrbValueFixnum(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString("42")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mruby.ToGo[int](value) != 42 {
		t.Fatalf("bad fixnum")
	}
}

func TestMrbValueString(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"foo"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "foo" {
		t.Fatalf("bad string")
	}
}

func TestMrbValueType(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	cases := []struct {
		Input    string
		Expected mruby.ValueType
	}{
		{
			`false`,
			mruby.TypeFalse,
		},
		// TypeFree - Type of value after GC collection
		{
			`true`,
			mruby.TypeTrue,
		},
		{
			`1`,
			mruby.TypeFixnum,
		},
		{
			`:test`,
			mruby.TypeSymbol,
		},
		// TypeUndef - Internal value used by mruby for undefined things (instance vars etc)
		// These all seem to get converted to exceptions before hitting userland
		{
			`1.1`,
			mruby.TypeFloat,
		},
		// TypeCptr
		{
			`Object.new`,
			mruby.TypeObject,
		},
		{
			`Object`,
			mruby.TypeClass,
		},
		{
			`module T; end; T`,
			mruby.TypeModule,
		},
		// TypeIClass
		// TypeSClass
		{
			`Proc.new { 1 }`,
			mruby.TypeProc,
		},
		{
			`[]`,
			mruby.TypeArray,
		},
		{
			`{}`,
			mruby.TypeHash,
		},
		{
			`"string"`,
			mruby.TypeString,
		},
		{
			`1..2`,
			mruby.TypeRange,
		},
		{
			`Exception.new`,
			mruby.TypeException,
		},
		// TypeFile
		// TypeEnv
		// TypeData
		// TypeFiber
		// TypeMaxDefine
		{
			`nil`,
			mruby.TypeNil,
		},
	}

	for _, tcase := range cases {
		r, err := mrb.LoadString(tcase.Input)
		if err != nil {
			t.Fatalf("loadstring failed for case %#v: %s", tcase, err)
		}
		if cType := r.Type(); cType != tcase.Expected {
			t.Fatalf("bad type: got %v, expected %v", cType, tcase.Expected)
		}
	}
}

func TestIntMrbValue(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value := mruby.ToRuby(mrb, 42)
	if mruby.ToGo[int](value) != 42 {
		t.Fatalf("bad value")
	}
}

func TestStringMrbValue(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value := mruby.ToRuby(mrb, "foo")
	if value.String() != "foo" {
		t.Fatalf("bad value")
	}
}

func TestValueClass(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	val, err := mrb.ObjectClass().New()
	if err != nil {
		t.Fatalf("Error constructing object of type Object: %v", err)
	}

	if !reflect.DeepEqual(val.Class(), mrb.ObjectClass()) {
		t.Fatal("Class of value was not equivalent to constructed class")
	}
}

func TestValueSingletonClass(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	fn := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		args := m.GetArgs()
		return mruby.ToRuby(mrb, mruby.ToGo[int](args[0])+mruby.ToGo[int](args[1])), nil
	}

	mrb.TopSelf().SingletonClass().DefineMethod("add", fn, mruby.ArgsReq(2))

	result, err := mrb.LoadString(`add(46, 2)`)
	if err != nil {
		t.Fatalf("Error parsing ruby code: %v", err)
	}

	if result.String() != "48" {
		t.Fatalf("Result %q was not equal to the target value of 48", result.String())
	}
}

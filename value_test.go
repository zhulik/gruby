package gruby_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zhulik/gruby"
)

func TestExceptionString_afterClose(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	parser := gruby.NewParser(mrb)
	defer parser.Close()
	context := gruby.NewCompileContext(mrb)
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

	var exc *gruby.ExceptionError
	errors.As(err, &exc)

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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"foo"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, err = value.Call("some_function_that_doesnt_exist")
	if err == nil {
		t.Fatalf("expected exception")
	}

	result, err := value.Call("==", gruby.ToRuby(mrb, "foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.Type() != gruby.TypeTrue {
		t.Fatalf("bad type")
	}
}

func TestMrbValueCallBlock(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"foo"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	block, err := mrb.LoadString(`Proc.new { |_| "bar" }`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := value.CallBlock("gsub", gruby.ToRuby(mrb, "foo"), block)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.Type() != gruby.TypeString {
		t.Fatalf("bad type")
	}
	if result.String() != "bar" {
		t.Fatalf("bad: %s", result)
	}
}

func TestMrbValueValue_impl(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	var _ gruby.Value = mrb.FalseValue()
}

func TestMrbValueFixnum(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString("42")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gruby.ToGo[int](value) != 42 {
		t.Fatalf("bad fixnum")
	}
}

func TestMrbValueString(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	cases := []struct {
		Input    string
		Expected gruby.ValueType
	}{
		{
			`false`,
			gruby.TypeFalse,
		},
		// TypeFree - Type of value after GC collection
		{
			`true`,
			gruby.TypeTrue,
		},
		{
			`1`,
			gruby.TypeFixnum,
		},
		{
			`:test`,
			gruby.TypeSymbol,
		},
		// TypeUndef - Internal value used by mruby for undefined things (instance vars etc)
		// These all seem to get converted to exceptions before hitting userland
		{
			`1.1`,
			gruby.TypeFloat,
		},
		// TypeCptr
		{
			`Object.new`,
			gruby.TypeObject,
		},
		{
			`Object`,
			gruby.TypeClass,
		},
		{
			`module T; end; T`,
			gruby.TypeModule,
		},
		// TypeIClass
		// TypeSClass
		{
			`Proc.new { 1 }`,
			gruby.TypeProc,
		},
		{
			`[]`,
			gruby.TypeArray,
		},
		{
			`{}`,
			gruby.TypeHash,
		},
		{
			`"string"`,
			gruby.TypeString,
		},
		{
			`1..2`,
			gruby.TypeRange,
		},
		{
			`Exception.new`,
			gruby.TypeException,
		},
		// TypeFile
		// TypeEnv
		// TypeData
		// TypeFiber
		// TypeMaxDefine
		{
			`nil`,
			gruby.TypeNil,
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value := gruby.ToRuby(mrb, 42)
	if gruby.ToGo[int](value) != 42 {
		t.Fatalf("bad value")
	}
}

func TestStringMrbValue(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value := gruby.ToRuby(mrb, "foo")
	if value.String() != "foo" {
		t.Fatalf("bad value")
	}
}

func TestValueClass(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	fn := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		args := m.GetArgs()
		return gruby.ToRuby(mrb, gruby.ToGo[int](args[0])+gruby.ToGo[int](args[1])), nil
	}

	mrb.TopSelf().SingletonClass().DefineMethod("add", fn, gruby.ArgsReq(2))

	result, err := mrb.LoadString(`add(46, 2)`)
	if err != nil {
		t.Fatalf("Error parsing ruby code: %v", err)
	}

	if result.String() != "48" {
		t.Fatalf("Result %q was not equal to the target value of 48", result.String())
	}
}

package mruby_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	mruby "github.com/zhulik/gruby"
)

func TestNewMrb(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	mrb.Close()
}

func TestMrbArena(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	idx := mrb.ArenaSave()
	mrb.ArenaRestore(idx)
}

func TestMrbModule(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	module := mrb.Module("Kernel")
	if module == nil {
		t.Fatal("module was nil and should not be")
	}
}

func TestMrbClass(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	class := mrb.Class("Object", nil)
	if class == nil {
		t.Fatal("class should not be nil")
	}

	mrb.DefineClass("Hello", mrb.ObjectClass())
	class = mrb.Class("Hello", mrb.ObjectClass())
	if class == nil {
		t.Fatal("class should not be nil")
	}
}

func TestMrbConstDefined(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	if !mrb.ConstDefined("Object", mrb.ObjectClass()) {
		t.Fatal("Object should be defined")
	}

	mrb.DefineClass("Hello", mrb.ObjectClass())
	if !mrb.ConstDefined("Hello", mrb.ObjectClass()) {
		t.Fatal("Hello should be defined")
	}
}

func TestMrbDefineClass(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	mrb.DefineClass("Hello", mrb.ObjectClass())
	_, err := mrb.LoadString("Hello")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mrb.DefineClass("World", nil)
	_, err = mrb.LoadString("World")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestMrbDefineClass_methodException(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	callback := func(mrb *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		value, err := mrb.LoadString(`raise "exception"`)
		if err != nil {
			var exc *mruby.ExceptionError
			errors.As(err, &exc)
			return nil, exc.Value
		}

		return value, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, mruby.ArgsNone())
	_, err := mrb.LoadString(`Hello.foo`)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestMrbDefineClassUnder(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	// Define an outer
	hello := mrb.DefineClass("Hello", mrb.ObjectClass())
	_, err := mrb.LoadString("Hello")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Inner
	mrb.DefineClassUnder("World", nil, hello)
	_, err = mrb.LoadString("Hello::World")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Inner defaults
	mrb.DefineClassUnder("Another", nil, nil)
	_, err = mrb.LoadString("Another")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestMrbDefineModule(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	mrb.DefineModule("Hello")
	_, err := mrb.LoadString("Hello")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestMrbDefineModuleUnder(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	// Define an outer
	hello := mrb.DefineModule("Hello")
	_, err := mrb.LoadString("Hello")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Inner
	mrb.DefineModuleUnder("World", hello)
	_, err = mrb.LoadString("Hello::World")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Inner defaults
	mrb.DefineModuleUnder("Another", nil)
	_, err = mrb.LoadString("Another")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestMrbFixnumValue(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value := mruby.ToRuby(mrb, 42)
	if value.Type() != mruby.TypeFixnum {
		t.Fatalf("should be fixnum")
	}
}

func TestMrbFullGC(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	aidx := mrb.ArenaSave()
	value := mruby.ToRuby(mrb, "foo")
	if value.IsDead() {
		t.Fatal("should not be dead")
	}

	mrb.ArenaRestore(aidx)
	mrb.FullGC()
	if !value.IsDead() {
		t.Fatal("should be dead")
	}
}

type testcase struct {
	args   string
	types  []mruby.ValueType
	result []string
}

func TestMrbGetArgs(t *testing.T) {
	t.Parallel()

	cases := []testcase{
		{
			`("foo")`,
			[]mruby.ValueType{mruby.TypeString},
			[]string{`"foo"`},
		},

		{
			`(true)`,
			[]mruby.ValueType{mruby.TypeTrue},
			[]string{`true`},
		},

		{
			`(Hello)`,
			[]mruby.ValueType{mruby.TypeClass},
			[]string{`Hello`},
		},

		{
			`() {}`,
			[]mruby.ValueType{mruby.TypeProc},
			nil,
		},

		{
			`(Hello, "bar", true)`,
			[]mruby.ValueType{mruby.TypeClass, mruby.TypeString, mruby.TypeTrue},
			[]string{`Hello`, `"bar"`, "true"},
		},

		{
			`("bar", true) {}`,
			[]mruby.ValueType{mruby.TypeString, mruby.TypeTrue, mruby.TypeProc},
			nil,
		},
	}

	// lots of this effort is centered around testing multithreaded behavior.

	for range 1000 {
		errChan := make(chan error, len(cases))

		for _, tcase := range cases {
			go func(tcase testcase) {
				var actual []mruby.Value
				testFunc := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
					actual = m.GetArgs()
					return self, nil
				}

				mrb := mruby.NewMrb()
				defer mrb.Close()
				class := mrb.DefineClass("Hello", mrb.ObjectClass())
				class.DefineClassMethod("test", testFunc, mruby.ArgsAny())
				_, err := mrb.LoadString("Hello.test" + tcase.args)
				if err != nil {
					errChan <- err
					return
				}

				if tcase.result != nil {
					if len(actual) != len(tcase.result) {
						errChan <- fmt.Errorf("%s: expected %d, got %d",
							tcase.args, len(tcase.result), len(actual))
						return
					}
				}

				actualStrings := make([]string, len(actual))
				actualTypes := make([]mruby.ValueType, len(actual))
				for idx, value := range actual {
					str, err := value.Call("inspect")
					if err != nil {
						errChan <- err
					}

					actualStrings[idx] = str.String()
					actualTypes[idx] = value.Type()
				}

				if !reflect.DeepEqual(actualTypes, tcase.types) {
					errChan <- fmt.Errorf("code: %s\nexpected: %#v\nactual: %#v",
						tcase.args, tcase.types, actualTypes)
					return
				}

				if tcase.result != nil {
					if !reflect.DeepEqual(actualStrings, tcase.result) {
						errChan <- fmt.Errorf("expected: %#v\nactual: %#v",
							tcase.result, actualStrings)
						return
					}
				}

				errChan <- nil
			}(tcase)
		}

		for range cases {
			if err := <-errChan; err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMrbGlobalVariable(t *testing.T) {
	t.Parallel()

	const (
		TestValue = "HELLO"
	)
	mrb := mruby.NewMrb()
	defer mrb.Close()
	if _, err := mrb.LoadString(fmt.Sprintf(`$a = "%s"`, TestValue)); err != nil {
		t.Fatalf("err: %s", err)
	}
	value := mrb.GetGlobalVariable("$a")
	if value.String() != TestValue {
		t.Fatalf("wrong value for $a: expected '%s', found '%s'", TestValue, value.String())
	}
	mrb.SetGlobalVariable("$b", mruby.ToRuby(mrb, TestValue))
	value, err := mrb.LoadString(`$b`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != TestValue {
		t.Fatalf("wrong value for $b: expected '%s', found '%s'", TestValue, value.String())
	}
}

func TestMrbInstanceVariable(t *testing.T) {
	t.Parallel()

	const (
		GoldenRetriever = "golden retriever"
		Husky           = "Husky"
	)
	mrb := mruby.NewMrb()
	defer mrb.Close()
	_, err := mrb.LoadString(`
		class Dog
			def initialize(breed)
				@breed = breed
			end
			def breed
				"cocker spaniel" # this line exists to ensure that it's not invoking the accessor method
			end
			def real_breed
				@breed
			end
		end
	`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	dogClass := mrb.Class("Dog", nil)
	if dogClass == nil {
		t.Fatalf("dog class not found")
	}
	inst, err := dogClass.New(mruby.ToRuby(mrb, GoldenRetriever))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	value := inst.GetInstanceVariable("@breed")
	if value.String() != GoldenRetriever {
		t.Fatalf("wrong value for Dog.@breed. expected: '%s', found: '%s'", GoldenRetriever, value.String())
	}
	inst.SetInstanceVariable("@breed", mruby.ToRuby(mrb, Husky))
	value = inst.GetInstanceVariable("@breed")
	if value.String() != Husky {
		t.Fatalf("wrong value for Dog.@breed. expected: '%s', found: '%s'", Husky, value.String())
	}
}

func TestMrbLoadString(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"HELLO"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value == nil {
		t.Fatalf("should have value")
	}
}

func TestMrbLoadString_twice(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"HELLO"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value == nil {
		t.Fatalf("should have value")
	}

	value, err = mrb.LoadString(`"WORLD"`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "WORLD" {
		t.Fatalf("bad: %s", value)
	}
}

func TestMrbLoadStringException(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	_, err := mrb.LoadString(`raise "An exception"`)

	if err == nil {
		t.Fatal("exception expected")
	}

	value, err := mrb.LoadString(`"test"`)
	if err != nil {
		t.Fatal("exception should have been cleared")
	}

	if value.String() != "test" {
		t.Fatal("bad test value returned")
	}
}

func TestMrbRaise(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	callback := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		return nil, m.GetArgs()[0]
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, mruby.ArgsReq(1))
	_, err := mrb.LoadString(`Hello.foo(ArgumentError.new("ouch"))`)
	if err == nil {
		t.Fatal("should have error")
	}
	if err.Error() != "ouch" {
		t.Fatalf("bad: %s", err)
	}
}

func TestMrbYield(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	callback := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		result, err := m.Yield(m.GetArgs()[0], mruby.ToRuby(mrb, 12), mruby.ToRuby(mrb, 30))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		return result, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, mruby.ArgsBlock())
	value, err := mrb.LoadString(`Hello.foo { |a, b| a + b }`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mruby.ToGo[int](value) != 42 {
		t.Fatalf("bad: %s", value)
	}
}

func TestMrbYieldException(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	callback := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		result, err := m.Yield(m.GetArgs()[0])
		if err != nil {
			var exc *mruby.ExceptionError
			errors.As(err, &exc)
			return nil, exc.Value
		}

		return result, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, mruby.ArgsBlock())
	_, err := mrb.LoadString(`Hello.foo { raise "exception" }`)
	if err == nil {
		t.Fatal("should error")
	}

	_, err = mrb.LoadString(`Hello.foo { 1 }`)
	if err != nil {
		t.Fatal("exception should have been cleared")
	}
}

func TestMrbRun(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	parser := mruby.NewParser(mrb)
	defer parser.Close()
	context := mruby.NewCompileContext(mrb)
	defer context.Close()

	_, err := parser.Parse(`
		if $do_raise
			raise "exception"
		else
			"rval"
		end`,
		context,
	)
	if err != nil {
		t.Fatal(err)
	}

	proc := parser.GenerateCode()

	// Enable proc exception raising & verify
	_, err = mrb.LoadString(`$do_raise = true`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mrb.Run(proc, nil)
	if err == nil {
		t.Fatalf("expected exception, %#v", err)
	}

	// Disable proc exception raising
	// If we still have an exception, it wasn't cleared from the previous invocation.
	_, err = mrb.LoadString(`$do_raise = false`)
	if err != nil {
		t.Fatal(err)
	}
	rval, err := mrb.Run(proc, nil)
	if err != nil {
		t.Fatalf("unexpected exception, %#v", err)
	}

	if rval.String() != "rval" {
		t.Fatalf("expected return value 'rval', got %#v", rval)
	}

	_, err = parser.Parse(`a = 10`, context)
	if err != nil {
		t.Fatal(err)
	}
	proc = parser.GenerateCode()

	stackKeep, _, err := mrb.RunWithContext(proc, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	if stackKeep != 2 {
		t.Fatalf("stack value was %d not 2; some variables may not have been captured", stackKeep)
	}

	_, err = parser.Parse(`a`, context)
	if err != nil {
		t.Fatal(err)
	}

	proc = parser.GenerateCode()

	var ret mruby.Value
	_, ret, err = mrb.RunWithContext(proc, nil, stackKeep)
	if err != nil {
		t.Fatal(err)
	}

	if ret.String() != "10" {
		t.Fatalf("Captured variable was not expected value: was %q", ret.String())
	}
}

func TestMrbDefineMethodConcurrent(t *testing.T) {
	t.Parallel()

	concurrency := 100
	numFuncs := 100

	callback := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		return m.GetArgs()[0], nil
	}

	syncChan := make(chan struct{}, concurrency)

	for range concurrency {
		go func() {
			mrb := mruby.NewMrb()
			defer mrb.Close()
			for i := range numFuncs {
				mrb.TopSelf().SingletonClass().DefineMethod(fmt.Sprintf("test%d", i), callback, mruby.ArgsAny())
			}

			syncChan <- struct{}{}
		}()
	}

	for range concurrency {
		<-syncChan
	}
}

func TestMrbStackedException(t *testing.T) {
	t.Parallel()

	var testClass *mruby.Class

	createException := func(m *mruby.Mrb, msg string) mruby.Value {
		val, err := m.Class("Exception", nil).New(mruby.ToRuby(m, msg))
		if err != nil {
			panic(fmt.Sprintf("could not construct exception for return: %v", err))
		}

		return val
	}

	testFunc := func(mrb *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		args := mrb.GetArgs()

		t, err := testClass.New()
		if err != nil {
			return nil, createException(mrb, err.Error())
		}

		v, err := t.Call("dotest!", args...)
		if err != nil {
			return nil, createException(mrb, err.Error())
		}

		return v, nil
	}

	doTestFunc := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		err := createException(m, "Fail us!")
		return nil, err
	}

	mrb := mruby.NewMrb()

	testClass = mrb.DefineClass("TestClass", nil)
	testClass.DefineMethod("dotest!", doTestFunc, mruby.ArgsReq(0)|mruby.ArgsOpt(3))

	mrb.TopSelf().SingletonClass().DefineMethod("test", testFunc, mruby.ArgsReq(0)|mruby.ArgsOpt(3))

	_, err := mrb.LoadString("test")
	if err == nil {
		t.Fatal("No exception when one was expected")
		return
	}

	mrb.Close()
	mrb = mruby.NewMrb()

	evalFunc := func(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
		arg := m.GetArgs()[0]
		_, err := self.CallBlock("instance_eval", arg)
		if err != nil {
			return nil, createException(m, err.Error())
		}

		return nil, nil
	}

	mrb.TopSelf().SingletonClass().DefineMethod("myeval", evalFunc, mruby.ArgsBlock())

	// TODO: fix me and enable back
	result, err := mrb.LoadString("myeval { raise 'foo' }")
	if err == nil {
		t.Fatal("did not error")
		return
	}

	if result != nil {
		t.Fatal("result was not cleared")
		return
	}

	mrb.Close()
}

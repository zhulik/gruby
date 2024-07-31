package gruby_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/zhulik/gruby"
)

func TestNewMrb(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	mrb.Close()
}

func TestMrbArena(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	idx := mrb.ArenaSave()
	mrb.ArenaRestore(idx)
}

func TestMrbModule(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	module := mrb.Module("Kernel")
	if module == nil {
		t.Fatal("module was nil and should not be")
	}
}

func TestMrbClass(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	callback := func(mrb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		value, err := mrb.LoadString(`raise "exception"`)
		if err != nil {
			var exc *gruby.ExceptionError
			errors.As(err, &exc)
			return nil, exc.Value
		}

		return value, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsNone())
	_, err := mrb.LoadString(`Hello.foo`)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestMrbDefineClassUnder(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	mrb.DefineModule("Hello")
	_, err := mrb.LoadString("Hello")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestMrbDefineModuleUnder(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value := gruby.ToRuby(mrb, 42)
	if value.Type() != gruby.TypeFixnum {
		t.Fatalf("should be fixnum")
	}
}

func TestMrbFullGC(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	aidx := mrb.ArenaSave()
	value := gruby.ToRuby(mrb, "foo")
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
	types  []gruby.ValueType
	result []string
}

func TestMrbGetArgs(t *testing.T) {
	t.Parallel()

	cases := []testcase{
		{
			`("foo")`,
			[]gruby.ValueType{gruby.TypeString},
			[]string{`"foo"`},
		},

		{
			`(true)`,
			[]gruby.ValueType{gruby.TypeTrue},
			[]string{`true`},
		},

		{
			`(Hello)`,
			[]gruby.ValueType{gruby.TypeClass},
			[]string{`Hello`},
		},

		{
			`() {}`,
			[]gruby.ValueType{gruby.TypeProc},
			nil,
		},

		{
			`(Hello, "bar", true)`,
			[]gruby.ValueType{gruby.TypeClass, gruby.TypeString, gruby.TypeTrue},
			[]string{`Hello`, `"bar"`, "true"},
		},

		{
			`("bar", true) {}`,
			[]gruby.ValueType{gruby.TypeString, gruby.TypeTrue, gruby.TypeProc},
			nil,
		},
	}

	// lots of this effort is centered around testing multithreaded behavior.

	for range 1000 {
		errChan := make(chan error, len(cases))

		for _, tcase := range cases {
			go func(tcase testcase) {
				var actual []gruby.Value
				testFunc := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
					actual = m.GetArgs()
					return self, nil
				}

				mrb := gruby.NewMrb()
				defer mrb.Close()
				class := mrb.DefineClass("Hello", mrb.ObjectClass())
				class.DefineClassMethod("test", testFunc, gruby.ArgsAny())
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
				actualTypes := make([]gruby.ValueType, len(actual))
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
	mrb := gruby.NewMrb()
	defer mrb.Close()
	if _, err := mrb.LoadString(fmt.Sprintf(`$a = "%s"`, TestValue)); err != nil {
		t.Fatalf("err: %s", err)
	}
	value := mrb.GetGlobalVariable("$a")
	if value.String() != TestValue {
		t.Fatalf("wrong value for $a: expected '%s', found '%s'", TestValue, value.String())
	}
	mrb.SetGlobalVariable("$b", gruby.ToRuby(mrb, TestValue))
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
	mrb := gruby.NewMrb()
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
	inst, err := dogClass.New(gruby.ToRuby(mrb, GoldenRetriever))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	value := inst.GetInstanceVariable("@breed")
	if value.String() != GoldenRetriever {
		t.Fatalf("wrong value for Dog.@breed. expected: '%s', found: '%s'", GoldenRetriever, value.String())
	}
	inst.SetInstanceVariable("@breed", gruby.ToRuby(mrb, Husky))
	value = inst.GetInstanceVariable("@breed")
	if value.String() != Husky {
		t.Fatalf("wrong value for Dog.@breed. expected: '%s', found: '%s'", Husky, value.String())
	}
}

func TestMrbLoadString(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	callback := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		return nil, m.GetArgs()[0]
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsReq(1))
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	callback := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		result, err := m.Yield(m.GetArgs()[0], gruby.ToRuby(mrb, 12), gruby.ToRuby(mrb, 30))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		return result, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsBlock())
	value, err := mrb.LoadString(`Hello.foo { |a, b| a + b }`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if gruby.ToGo[int](value) != 42 {
		t.Fatalf("bad: %s", value)
	}
}

func TestMrbYieldException(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	callback := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		result, err := m.Yield(m.GetArgs()[0])
		if err != nil {
			var exc *gruby.ExceptionError
			errors.As(err, &exc)
			return nil, exc.Value
		}

		return result, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsBlock())
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

	mrb := gruby.NewMrb()
	defer mrb.Close()

	parser := gruby.NewParser(mrb)
	defer parser.Close()
	context := gruby.NewCompileContext(mrb)
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

	var ret gruby.Value
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

	callback := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		return m.GetArgs()[0], nil
	}

	syncChan := make(chan struct{}, concurrency)

	for range concurrency {
		go func() {
			mrb := gruby.NewMrb()
			defer mrb.Close()
			for i := range numFuncs {
				mrb.TopSelf().SingletonClass().DefineMethod(fmt.Sprintf("test%d", i), callback, gruby.ArgsAny())
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

	var testClass *gruby.Class

	createException := func(m *gruby.GRuby, msg string) gruby.Value {
		val, err := m.Class("Exception", nil).New(gruby.ToRuby(m, msg))
		if err != nil {
			panic(fmt.Sprintf("could not construct exception for return: %v", err))
		}

		return val
	}

	testFunc := func(mrb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
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

	doTestFunc := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		err := createException(m, "Fail us!")
		return nil, err
	}

	mrb := gruby.NewMrb()

	testClass = mrb.DefineClass("TestClass", nil)
	testClass.DefineMethod("dotest!", doTestFunc, gruby.ArgsReq(0)|gruby.ArgsOpt(3))

	mrb.TopSelf().SingletonClass().DefineMethod("test", testFunc, gruby.ArgsReq(0)|gruby.ArgsOpt(3))

	_, err := mrb.LoadString("test")
	if err == nil {
		t.Fatal("No exception when one was expected")
		return
	}

	mrb.Close()
	mrb = gruby.NewMrb()

	evalFunc := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		arg := m.GetArgs()[0]
		_, err := self.CallBlock("instance_eval", arg)
		if err != nil {
			return nil, createException(m, err.Error())
		}

		return nil, nil
	}

	mrb.TopSelf().SingletonClass().DefineMethod("myeval", evalFunc, gruby.ArgsBlock())

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

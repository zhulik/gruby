package gruby_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func TestArena(t *testing.T) {
	t.Parallel()

	grb := must(gruby.New())
	defer grb.Close()

	idx := grb.ArenaSave()
	grb.ArenaRestore(idx)
}

func TestModule(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	module := grb.Module("Kernel")
	g.Expect(module).ToNot(BeNil())
}

func TestClass(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	class := grb.Class("Object", nil)
	g.Expect(class).ToNot(BeNil())

	grb.DefineClass("Hello", grb.ObjectClass())
	class = grb.Class("Hello", grb.ObjectClass())
	g.Expect(class).ToNot(BeNil())
}

func TestConstDefined(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	g.Expect(grb.ConstDefined("Object", grb.ObjectClass())).To(BeTrue())

	grb.DefineClass("Hello", grb.ObjectClass())
	g.Expect(grb.ConstDefined("Hello", grb.ObjectClass())).To(BeTrue())
}

func TestDefineClass(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	grb.DefineClass("Hello", grb.ObjectClass())
	_, err := grb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())

	grb.DefineClass("World", nil)
	_, err = grb.LoadString("World")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestDefineClass_methodException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	callback := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		value, err := grb.LoadString(`raise "exception"`)
		if err != nil {
			var exc *gruby.ExceptionError
			errors.As(err, &exc)
			return nil, exc.Value
		}

		return value, nil
	}

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsNone())
	_, err := grb.LoadString(`Hello.foo`)
	g.Expect(err).To(HaveOccurred())
}

func TestDefineClassUnder(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	// Define an outer
	hello := grb.DefineClass("Hello", grb.ObjectClass())
	_, err := grb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner
	grb.DefineClassUnder("World", nil, hello)
	_, err = grb.LoadString("Hello::World")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner defaults
	grb.DefineClassUnder("Another", nil, nil)
	_, err = grb.LoadString("Another")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestDefineModule(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	grb.DefineModule("Hello")
	_, err := grb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestDefineModuleUnder(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	// Define an outer
	hello := grb.DefineModule("Hello")
	_, err := grb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner
	grb.DefineModuleUnder("World", hello)
	_, err = grb.LoadString("Hello::World")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner defaults
	grb.DefineModuleUnder("Another", nil)
	_, err = grb.LoadString("Another")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestFixnumValue(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value := gruby.ToRuby(grb, 42)
	g.Expect(value.Type()).To(Equal(gruby.TypeFixnum))
}

func TestFullGC(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	aidx := grb.ArenaSave()
	value := gruby.ToRuby(grb, "foo")
	g.Expect(value.IsDead()).To(BeFalse())

	grb.ArenaRestore(aidx)
	grb.FullGC()
	g.Expect(value.IsDead()).To(BeTrue())
}

type testcase struct {
	args   string
	types  []gruby.ValueType
	result []string
}

func TestGetArgs(t *testing.T) {
	t.Parallel()
	g := NewG(t)

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
				testFunc := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
					actual = grb.GetArgs()
					return self, nil
				}

				grb := must(gruby.New())
				defer grb.Close()
				class := grb.DefineClass("Hello", grb.ObjectClass())
				class.DefineClassMethod("test", testFunc, gruby.ArgsAny())
				_, err := grb.LoadString("Hello.test" + tcase.args)
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
			g.Expect(<-errChan).ToNot(HaveOccurred())
		}
	}
}

func TestGlobalVariable(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	const (
		TestValue = "HELLO"
	)
	grb := must(gruby.New())
	defer grb.Close()
	_, err := grb.LoadString(fmt.Sprintf(`$a = "%s"`, TestValue))
	g.Expect(err).ToNot(HaveOccurred())

	value := grb.GetGlobalVariable("$a")
	g.Expect(value.String()).To(Equal(TestValue))

	grb.SetGlobalVariable("$b", gruby.ToRuby(grb, TestValue))
	value, err = grb.LoadString(`$b`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal(TestValue))
}

func TestInstanceVariable(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	const (
		GoldenRetriever = "golden retriever"
		Husky           = "Husky"
	)
	grb := must(gruby.New())
	defer grb.Close()
	_, err := grb.LoadString(`
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
	g.Expect(err).ToNot(HaveOccurred())
	dogClass := grb.Class("Dog", nil)
	g.Expect(dogClass).ToNot(BeNil())

	inst, err := dogClass.New(gruby.ToRuby(grb, GoldenRetriever))
	g.Expect(err).ToNot(HaveOccurred())

	value := inst.GetInstanceVariable("@breed")
	g.Expect(value.String()).To(Equal(GoldenRetriever))

	inst.SetInstanceVariable("@breed", gruby.ToRuby(grb, Husky))
	value = inst.GetInstanceVariable("@breed")
	g.Expect(value.String()).To(Equal(Husky))
}

func TestLoadString(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString(`"HELLO"`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value).ToNot(BeNil())
}

func TestLoadString_twice(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString(`"HELLO"`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value).ToNot(BeNil())

	value, err = grb.LoadString(`"WORLD"`)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(value.String()).To(Equal("WORLD"))
}

func TestLoadStringException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	_, err := grb.LoadString(`raise "An exception"`)

	g.Expect(err).To(HaveOccurred())

	value, err := grb.LoadString(`"test"`)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(value.String()).To(Equal("test"))
}

func TestRaise(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	callback := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		return nil, grb.GetArgs()[0]
	}

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsReq(1))
	_, err := grb.LoadString(`Hello.foo(ArgumentError.new("ouch"))`)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("ouch"))
}

func TestYield(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	callback := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		result, err := grb.Yield(grb.GetArgs()[0], gruby.ToRuby(grb, 12), gruby.ToRuby(grb, 30))
		g.Expect(err).ToNot(HaveOccurred())

		return result, nil
	}

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsBlock())
	value, err := grb.LoadString(`Hello.foo { |a, b| a + b }`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(gruby.ToGo[int](value)).To(Equal(42))
}

func TestYieldException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	callback := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		result, err := grb.Yield(grb.GetArgs()[0])
		if err != nil {
			var exc *gruby.ExceptionError
			errors.As(err, &exc)
			return nil, exc.Value
		}

		return result, nil
	}

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsBlock())
	_, err := grb.LoadString(`Hello.foo { raise "exception" }`)
	g.Expect(err).To(HaveOccurred())

	_, err = grb.LoadString(`Hello.foo { 1 }`)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestRun(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	parser := gruby.NewParser(grb)
	defer parser.Close()
	context := gruby.NewCompileContext(grb)
	defer context.Close()

	_, err := parser.Parse(`
		if $do_raise
			raise "exception"
		else
			"rval"
		end`,
		context,
	)
	g.Expect(err).ToNot(HaveOccurred())

	proc := parser.GenerateCode()

	// Enable proc exception raising & verify
	_, err = grb.LoadString(`$do_raise = true`)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = grb.Run(proc, nil)
	g.Expect(err).To(HaveOccurred())

	// Disable proc exception raising
	// If we still have an exception, it wasn't cleared from the previous invocation.
	_, err = grb.LoadString(`$do_raise = false`)
	g.Expect(err).ToNot(HaveOccurred())

	rval, err := grb.Run(proc, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(rval.String()).To(Equal("rval"))

	_, err = parser.Parse(`a = 10`, context)
	g.Expect(err).ToNot(HaveOccurred())
	proc = parser.GenerateCode()

	stackKeep, _, err := grb.RunWithContext(proc, nil, 0)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(stackKeep).To(Equal(2), "some variables may not have been captured")

	_, err = parser.Parse(`a`, context)
	g.Expect(err).ToNot(HaveOccurred())

	proc = parser.GenerateCode()

	var ret gruby.Value
	_, ret, err = grb.RunWithContext(proc, nil, stackKeep)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(ret.String()).To(Equal("10"))
}

func TestDefineMethodConcurrent(t *testing.T) {
	t.Parallel()

	concurrency := 100
	numFuncs := 100

	callback := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		return grb.GetArgs()[0], nil
	}

	syncChan := make(chan struct{}, concurrency)

	for range concurrency {
		go func() {
			grb := must(gruby.New())
			defer grb.Close()
			for i := range numFuncs {
				grb.TopSelf().SingletonClass().DefineMethod(fmt.Sprintf("test%d", i), callback, gruby.ArgsAny())
			}

			syncChan <- struct{}{}
		}()
	}

	for range concurrency {
		<-syncChan
	}
}

func TestStackedException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	var testClass *gruby.Class

	createException := func(grb *gruby.GRuby, msg string) gruby.Value {
		val, err := grb.Class("Exception", nil).New(gruby.ToRuby(grb, msg))
		g.Expect(err).ToNot(HaveOccurred())
		return val
	}

	testFunc := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		args := grb.GetArgs()

		t, err := testClass.New()
		if err != nil {
			return nil, createException(grb, err.Error())
		}

		v, err := t.Call("dotest!", args...)
		if err != nil {
			return nil, createException(grb, err.Error())
		}

		return v, nil
	}

	doTestFunc := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		err := createException(grb, "Fail us!")
		return nil, err
	}

	grb := must(gruby.New())

	testClass = grb.DefineClass("TestClass", nil)
	testClass.DefineMethod("dotest!", doTestFunc, gruby.ArgsReq(0)|gruby.ArgsOpt(3))

	grb.TopSelf().SingletonClass().DefineMethod("test", testFunc, gruby.ArgsReq(0)|gruby.ArgsOpt(3))

	_, err := grb.LoadString("test")
	g.Expect(err).To(HaveOccurred())

	grb.Close()
	grb = must(gruby.New())

	evalFunc := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		arg := grb.GetArgs()[0]
		_, err := self.CallBlock("instance_eval", arg)
		if err != nil {
			return nil, createException(grb, err.Error())
		}

		return nil, nil
	}

	grb.TopSelf().SingletonClass().DefineMethod("myeval", evalFunc, gruby.ArgsBlock())

	result, err := grb.LoadString("myeval { raise 'foo' }")
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(BeNil())

	grb.Close()
}

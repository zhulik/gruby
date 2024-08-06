package gruby_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestMrbArena(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	idx := mrb.ArenaSave()
	mrb.ArenaRestore(idx)
}

func TestMrbModule(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	module := mrb.Module("Kernel")
	g.Expect(module).ToNot(BeNil())
}

func TestMrbClass(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	class := mrb.Class("Object", nil)
	g.Expect(class).ToNot(BeNil())

	mrb.DefineClass("Hello", mrb.ObjectClass())
	class = mrb.Class("Hello", mrb.ObjectClass())
	g.Expect(class).ToNot(BeNil())
}

func TestMrbConstDefined(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	g.Expect(mrb.ConstDefined("Object", mrb.ObjectClass())).To(BeTrue())

	mrb.DefineClass("Hello", mrb.ObjectClass())
	g.Expect(mrb.ConstDefined("Hello", mrb.ObjectClass())).To(BeTrue())
}

func TestMrbDefineClass(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	mrb.DefineClass("Hello", mrb.ObjectClass())
	_, err := mrb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())

	mrb.DefineClass("World", nil)
	_, err = mrb.LoadString("World")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMrbDefineClass_methodException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

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
	g.Expect(err).To(HaveOccurred())
}

func TestMrbDefineClassUnder(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	// Define an outer
	hello := mrb.DefineClass("Hello", mrb.ObjectClass())
	_, err := mrb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner
	mrb.DefineClassUnder("World", nil, hello)
	_, err = mrb.LoadString("Hello::World")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner defaults
	mrb.DefineClassUnder("Another", nil, nil)
	_, err = mrb.LoadString("Another")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMrbDefineModule(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	mrb.DefineModule("Hello")
	_, err := mrb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMrbDefineModuleUnder(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	// Define an outer
	hello := mrb.DefineModule("Hello")
	_, err := mrb.LoadString("Hello")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner
	mrb.DefineModuleUnder("World", hello)
	_, err = mrb.LoadString("Hello::World")
	g.Expect(err).ToNot(HaveOccurred())

	// Inner defaults
	mrb.DefineModuleUnder("Another", nil)
	_, err = mrb.LoadString("Another")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMrbFixnumValue(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value := gruby.ToRuby(mrb, 42)
	g.Expect(value.Type()).To(Equal(gruby.TypeFixnum))
}

func TestMrbFullGC(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	aidx := mrb.ArenaSave()
	value := gruby.ToRuby(mrb, "foo")
	g.Expect(value.IsDead()).To(BeFalse())

	mrb.ArenaRestore(aidx)
	mrb.FullGC()
	g.Expect(value.IsDead()).To(BeTrue())
}

type testcase struct {
	args   string
	types  []gruby.ValueType
	result []string
}

func TestMrbGetArgs(t *testing.T) {
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
			g.Expect(<-errChan).ToNot(HaveOccurred())
		}
	}
}

func TestMrbGlobalVariable(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	const (
		TestValue = "HELLO"
	)
	mrb := gruby.NewMrb()
	defer mrb.Close()
	_, err := mrb.LoadString(fmt.Sprintf(`$a = "%s"`, TestValue))
	g.Expect(err).ToNot(HaveOccurred())

	value := mrb.GetGlobalVariable("$a")
	g.Expect(value.String()).To(Equal(TestValue))

	mrb.SetGlobalVariable("$b", gruby.ToRuby(mrb, TestValue))
	value, err = mrb.LoadString(`$b`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal(TestValue))
}

func TestMrbInstanceVariable(t *testing.T) {
	t.Parallel()
	g := NewG(t)

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
	g.Expect(err).ToNot(HaveOccurred())
	dogClass := mrb.Class("Dog", nil)
	g.Expect(dogClass).ToNot(BeNil())

	inst, err := dogClass.New(gruby.ToRuby(mrb, GoldenRetriever))
	g.Expect(err).ToNot(HaveOccurred())

	value := inst.GetInstanceVariable("@breed")
	g.Expect(value.String()).To(Equal(GoldenRetriever))

	inst.SetInstanceVariable("@breed", gruby.ToRuby(mrb, Husky))
	value = inst.GetInstanceVariable("@breed")
	g.Expect(value.String()).To(Equal(Husky))
}

func TestMrbLoadString(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"HELLO"`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value).ToNot(BeNil())
}

func TestMrbLoadString_twice(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`"HELLO"`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value).ToNot(BeNil())

	value, err = mrb.LoadString(`"WORLD"`)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(value.String()).To(Equal("WORLD"))
}

func TestMrbLoadStringException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	_, err := mrb.LoadString(`raise "An exception"`)

	g.Expect(err).To(HaveOccurred())

	value, err := mrb.LoadString(`"test"`)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(value.String()).To(Equal("test"))
}

func TestMrbRaise(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	callback := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		return nil, m.GetArgs()[0]
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsReq(1))
	_, err := mrb.LoadString(`Hello.foo(ArgumentError.new("ouch"))`)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("ouch"))
}

func TestMrbYield(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	callback := func(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		result, err := m.Yield(m.GetArgs()[0], gruby.ToRuby(mrb, 12), gruby.ToRuby(mrb, 30))
		g.Expect(err).ToNot(HaveOccurred())

		return result, nil
	}

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", callback, gruby.ArgsBlock())
	value, err := mrb.LoadString(`Hello.foo { |a, b| a + b }`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(gruby.ToGo[int](value)).To(Equal(42))
}

func TestMrbYieldException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

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
	g.Expect(err).To(HaveOccurred())

	_, err = mrb.LoadString(`Hello.foo { 1 }`)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMrbRun(t *testing.T) {
	t.Parallel()
	g := NewG(t)

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
	g.Expect(err).ToNot(HaveOccurred())

	proc := parser.GenerateCode()

	// Enable proc exception raising & verify
	_, err = mrb.LoadString(`$do_raise = true`)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = mrb.Run(proc, nil)
	g.Expect(err).To(HaveOccurred())

	// Disable proc exception raising
	// If we still have an exception, it wasn't cleared from the previous invocation.
	_, err = mrb.LoadString(`$do_raise = false`)
	g.Expect(err).ToNot(HaveOccurred())

	rval, err := mrb.Run(proc, nil)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(rval.String()).To(Equal("rval"))

	_, err = parser.Parse(`a = 10`, context)
	g.Expect(err).ToNot(HaveOccurred())
	proc = parser.GenerateCode()

	stackKeep, _, err := mrb.RunWithContext(proc, nil, 0)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(stackKeep).To(Equal(2), "some variables may not have been captured")

	_, err = parser.Parse(`a`, context)
	g.Expect(err).ToNot(HaveOccurred())

	proc = parser.GenerateCode()

	var ret gruby.Value
	_, ret, err = mrb.RunWithContext(proc, nil, stackKeep)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(ret.String()).To(Equal("10"))
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
	g := NewG(t)

	var testClass *gruby.Class

	createException := func(m *gruby.GRuby, msg string) gruby.Value {
		val, err := m.Class("Exception", nil).New(gruby.ToRuby(m, msg))
		g.Expect(err).ToNot(HaveOccurred())
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
	g.Expect(err).To(HaveOccurred())

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
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(BeNil())

	mrb.Close()
}

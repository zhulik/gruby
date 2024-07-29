package mruby

import (
	"testing"
)

func TestClassDefineClassMethod(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", testCallback, ArgsNone())
	value, err := mrb.LoadString("Hello.foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	testCallbackResult(t, value)
}

func TestClassDefineConst(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineConst("FOO", ToRuby(mrb, "bar"))
	value, err := mrb.LoadString("Hello::FOO")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "bar" {
		t.Fatalf("bad: %s", value)
	}
}

func TestClassDefineMethod(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineMethod("foo", testCallback, ArgsNone())
	value, err := mrb.LoadString("Hello.new.foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	testCallbackResult(t, value)
}

func TestClassNew(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineMethod("foo", testCallback, ArgsNone())

	instance, err := class.New()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	value, err := instance.Call("foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	testCallbackResult(t, value)
}

func TestClassNewException(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()

	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineMethod("initialize", testCallbackException, ArgsNone())

	_, err := class.New()
	if err == nil {
		t.Fatalf("expected exception")
	}

	// Verify exception is cleared
	val, err := mrb.LoadString(`"test"`)
	if err != nil {
		t.Fatalf("unexpected exception: %#v", err)
	}

	if val.String() != "test" {
		t.Fatalf("expected val 'test', got %#v", val)
	}
}

func TestClassValue(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	if class.Type() != TypeClass {
		t.Fatalf("bad: %d", class.Type())
	}
}

package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestClassDefineClassMethod(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineClassMethod("foo", testCallback, gruby.ArgsNone())
	value, err := mrb.LoadString("Hello.foo")
	g.Expect(err).ToNot(HaveOccurred())

	testCallbackResult(g, value)
}

func TestClassDefineConst(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineConst("FOO", gruby.ToRuby(mrb, "bar"))
	value, err := mrb.LoadString("Hello::FOO")

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("bar"))
}

func TestClassDefineMethod(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineMethod("foo", testCallback, gruby.ArgsNone())
	value, err := mrb.LoadString("Hello.new.foo")
	g.Expect(err).ToNot(HaveOccurred())

	testCallbackResult(g, value)
}

func TestClassNew(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineMethod("foo", testCallback, gruby.ArgsNone())

	instance, err := class.New()
	g.Expect(err).ToNot(HaveOccurred())

	value, err := instance.Call("foo")
	g.Expect(err).ToNot(HaveOccurred())

	testCallbackResult(g, value)
}

func TestClassNewException(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()

	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	class.DefineMethod("initialize", testCallbackException, gruby.ArgsNone())

	_, err := class.New()
	g.Expect(err).To(HaveOccurred())

	// Verify exception is cleared
	val, err := mrb.LoadString(`"test"`)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(val.String()).To(Equal("test"))
}

func TestClassValue(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	class := mrb.DefineClass("Hello", mrb.ObjectClass())
	g.Expect(class.Type()).To(Equal(class.Type()))
}

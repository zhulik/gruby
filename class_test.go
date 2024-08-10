package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestClassDefineClassMethod(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())
	defer grb.Close()

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineClassMethod("foo", testCallback, gruby.ArgsNone())
	value, err := grb.LoadString("Hello.foo")
	g.Expect(err).ToNot(HaveOccurred())

	testCallbackResult(g, value)
}

func TestClassDefineConst(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())
	defer grb.Close()

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineConst("FOO", gruby.MustToRuby(grb, "bar"))
	value, err := grb.LoadString("Hello::FOO")

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("bar"))
}

func TestClassDefineMethod(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())
	defer grb.Close()

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineMethod("foo", testCallback, gruby.ArgsNone())
	value, err := grb.LoadString("Hello.new.foo")
	g.Expect(err).ToNot(HaveOccurred())

	testCallbackResult(g, value)
}

func TestClassNew(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())
	defer grb.Close()

	class := grb.DefineClass("Hello", grb.ObjectClass())
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

	grb := gruby.Must(gruby.New())

	defer grb.Close()

	class := grb.DefineClass("Hello", grb.ObjectClass())
	class.DefineMethod("initialize", testCallbackException, gruby.ArgsNone())

	_, err := class.New()
	g.Expect(err).To(HaveOccurred())

	// Verify exception is cleared
	val, err := grb.LoadString(`"test"`)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(val.String()).To(Equal("test"))
}

func TestClassValue(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())
	defer grb.Close()

	class := grb.DefineClass("Hello", grb.ObjectClass())
	g.Expect(class.Type()).To(Equal(class.Type()))
}

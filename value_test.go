package gruby_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestExceptionString_afterClose(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	_, err := grb.LoadString(`clearly a syntax error`)
	grb.Close()
	// This panics before the bug fix that this test tests
	g.Expect(err.Error()).To(Equal("undefined method 'error'"))
}

func TestExceptionBacktrace(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	parser := gruby.NewParser(grb)
	defer parser.Close()
	context := gruby.NewCompileContext(grb)
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
	g.Expect(err).ToNot(HaveOccurred())

	proc := parser.GenerateCode()
	_, err = grb.Run(proc, nil)
	g.Expect(err).To(HaveOccurred())

	var exc *gruby.ExceptionError
	errors.As(err, &exc)

	g.Expect(exc.Message).To(Equal("Exception"))
	g.Expect(exc.File).To(Equal("hello.rb"))
	g.Expect(exc.Line).To(Equal(3))
	g.Expect(exc.Backtrace).To(HaveLen(4))
}

func TestValueCall(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString(`"foo"`)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = value.Call("some_function_that_doesnt_exist")
	g.Expect(err).To(HaveOccurred())

	result, err := value.Call("==", gruby.MustToRuby(grb, "foo"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.Type()).To(Equal(gruby.TypeTrue))
}

func TestValueCallBlock(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString(`"foo"`)
	g.Expect(err).ToNot(HaveOccurred())

	block, err := grb.LoadString(`Proc.new { |_| "bar" }`)
	g.Expect(err).ToNot(HaveOccurred())

	result, err := value.CallBlock("gsub", gruby.MustToRuby(grb, "foo"), block)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(result.Type()).To(Equal(gruby.TypeString))
	g.Expect(result.String()).To(Equal("bar"))
}

func TestValueFixnum(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString("42")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(gruby.MustToGo[int](value)).To(Equal(42))
}

func TestValueString(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString(`"foo"`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("foo"))
}

func TestValueType(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

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
		r, err := grb.LoadString(tcase.Input)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(r.Type()).To(Equal(tcase.Expected))
	}
}

func TestIntGValue(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value := gruby.MustToRuby(grb, 42)
	g.Expect(gruby.MustToGo[int](value)).To(Equal(42))
}

func TestStringGValue(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value := gruby.MustToRuby(grb, "foo")
	g.Expect(value.String()).To(Equal("foo"))
}

func TestValueClass(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	val, err := grb.ObjectClass().New()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(val.Class()).To(Equal(grb.ObjectClass()))
}

func TestValueSingletonClass(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	fn := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		args := grb.GetArgs()
		return gruby.MustToRuby(grb, gruby.MustToGo[int](args[0])+gruby.MustToGo[int](args[1])), nil
	}

	grb.TopSelf().SingletonClass().DefineMethod("add", fn, gruby.ArgsReq(2))

	result, err := grb.LoadString(`add(46, 2)`)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.String()).To(Equal("48"))
}

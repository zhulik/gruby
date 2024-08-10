package gruby_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestDecode(t *testing.T) {
	t.Parallel()
	g := NewG(t)
	grb := gruby.Must(gruby.New())
	defer grb.Close()

	type structString struct {
		Foo string
	}

	var outBool bool
	var outFloat64 float64
	var outInt int
	var outMap, outMap2 map[string]string
	var outPtrInt *int
	var outSlice []string
	var outString string
	var outStructString structString

	cases := []struct {
		Input    string
		Output   interface{}
		Expected interface{}
	}{
		// Booleans
		{
			"true",
			&outBool,
			true,
		},

		{
			"false",
			&outBool,
			false,
		},

		// Float
		{
			"1.2",
			&outFloat64,
			float64(1.2000000476837158),
		},

		// Int
		{
			"32",
			&outInt,
			int(32),
		},

		{
			`"32"`,
			&outInt,
			int(32),
		},

		// Map
		{
			`{"foo" => "bar"}`,
			&outMap,
			map[string]string{"foo": "bar"},
		},

		{
			`{32 => "bar"}`,
			&outMap2,
			map[string]string{"32": "bar"},
		},

		// Slice
		{
			`["foo", "bar"]`,
			&outSlice,
			[]string{"foo", "bar"},
		},

		// Ptr
		{
			`32`,
			&outPtrInt,
			32,
		},

		// String
		{
			`32`,
			&outString,
			"32",
		},

		{
			`"32"`,
			&outString,
			"32",
		},

		// Struct from Hash
		{
			`{"foo" => "bar"}`,
			&outStructString,
			structString{Foo: "bar"},
		},

		// Struct from object with methods
		{
			testDecodeObjectMethods,
			&outStructString,
			structString{Foo: "bar"},
		},
	}
	for _, tcase := range cases {
		value, err := grb.LoadString(tcase.Input)
		g.Expect(err).ToNot(HaveOccurred())

		err = gruby.Decode(tcase.Output, value)
		g.Expect(err).ToNot(HaveOccurred())

		val := reflect.ValueOf(tcase.Output)
		for val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		g.Expect(val.Interface()).To(Equal(tcase.Expected))
	}
}

func TestDecodeInterface(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	cases := []struct {
		Input    string
		Expected interface{}
	}{
		// Booleans
		{
			"true",
			true,
		},

		{
			"false",
			false,
		},

		// Float
		{
			"1.2",
			float64(1.2000000476837158),
		},

		// Int
		{
			"32",
			int(32),
		},

		// Map
		{
			`{"foo" => "bar"}`,
			map[string]interface{}{"foo": "bar"},
		},

		{
			`{32 => "bar"}`,
			map[string]interface{}{"32": "bar"},
		},

		// Slice
		{
			`["foo", "bar"]`,
			[]interface{}{"foo", "bar"},
		},

		// String
		{
			`"32"`,
			"32",
		},
	}

	for _, tcase := range cases {
		grb := gruby.Must(gruby.New())
		value, err := grb.LoadString(tcase.Input)
		g.Expect(err).ToNot(HaveOccurred())

		var result interface{}
		err = gruby.Decode(&result, value)
		grb.Close()

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(result).To(Equal(tcase.Expected))
	}
}

const testDecodeObjectMethods = `
class Foo
	def foo
		"bar"
	end
end

Foo.new
`

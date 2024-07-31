package gruby_test

import (
	"reflect"
	"testing"

	"github.com/zhulik/gruby"
)

func TestDecode(t *testing.T) {
	t.Parallel()

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
		mrb := gruby.NewMrb()
		value, err := mrb.LoadString(tcase.Input)
		if err != nil {
			mrb.Close()
			t.Fatalf("err: %s\n\n%s", err, tcase.Input)
		}

		err = gruby.Decode(tcase.Output, value)
		mrb.Close()
		if err != nil {
			t.Fatalf("input=%s output=%+v err: %s", tcase.Input, tcase.Output, err)
		}

		val := reflect.ValueOf(tcase.Output)
		for val.Kind() == reflect.Ptr {
			val = reflect.Indirect(val)
		}
		actual := val.Interface()
		if !reflect.DeepEqual(actual, tcase.Expected) {
			t.Fatalf("bad: %#v\n\n%#v", actual, tcase.Expected)
		}
	}
}

func TestDecodeInterface(t *testing.T) {
	t.Parallel()

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
		mrb := gruby.NewMrb()
		value, err := mrb.LoadString(tcase.Input)
		if err != nil {
			mrb.Close()
			t.Fatalf("err: %s\n\n%s", err, tcase.Input)
		}

		var result interface{}
		err = gruby.Decode(&result, value)
		mrb.Close()
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(result, tcase.Expected) {
			t.Fatalf("bad: \n\n%s\n\n%#v\n\n%#v", tcase.Input, result, tcase.Expected)
		}
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

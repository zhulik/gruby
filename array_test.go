package gruby_test

import (
	"testing"

	"github.com/zhulik/gruby"
)

func TestArray(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`["foo", "bar", "baz", false]`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	array := gruby.ToGo[*gruby.Array](value)

	// Len
	if n := array.Len(); n != 4 {
		t.Fatalf("bad: %d", n)
	}

	// Get
	value, err = array.Get(1)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "bar" {
		t.Fatalf("bad: %s", value)
	}

	// Get bool
	value, err = array.Get(3)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if valType := value.Type(); valType != gruby.TypeFalse {
		t.Fatalf("bad type: %v", valType)
	}
	if value.String() != "false" {
		t.Fatalf("bad: %s", value)
	}
}

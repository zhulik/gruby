package mruby

import (
	"testing"
)

func TestHash(t *testing.T) {
	t.Parallel()

	mrb := NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`{"foo" => "bar", "baz" => false}`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	hash := ToGo[*Hash](value)

	// Get
	value, err = hash.Get(ToRuby(mrb, "foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "bar" {
		t.Fatalf("bad: %s", value)
	}

	// Get false type
	value, err = hash.Get(ToRuby(mrb, "baz"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if valType := value.Type(); valType != TypeFalse {
		t.Fatalf("bad type: %v", valType)
	}
	if value.String() != "false" {
		t.Fatalf("bad: %s", value)
	}

	// Set
	err = hash.Set(ToRuby(mrb, "foo"), ToRuby(mrb, "baz"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	value, err = hash.Get(ToRuby(mrb, "foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "baz" {
		t.Fatalf("bad: %s", value)
	}

	// Keys
	value, err = hash.Keys()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.Type() != TypeArray {
		t.Fatalf("bad: %v", value.Type())
	}
	if value.String() != `["foo", "baz"]` {
		t.Fatalf("bad: %s", value)
	}

	// Delete
	value = hash.Delete(ToRuby(mrb, "foo"))
	if value.String() != "baz" {
		t.Fatalf("bad: %s", value)
	}

	value, err = hash.Keys()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != `["baz"]` {
		t.Fatalf("bad: %s", value)
	}

	// Delete non-existing
	value = hash.Delete(ToRuby(mrb, "nope"))
	if value != nil {
		t.Fatalf("bad: %s", value)
	}
}

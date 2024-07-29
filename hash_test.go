package mruby_test

import (
	"testing"

	mruby "github.com/zhulik/gruby"
)

func TestHash(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	value, err := mrb.LoadString(`{"foo" => "bar", "baz" => false}`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	hash := mruby.ToGo[*mruby.Hash](value)

	// Get
	value, err = hash.Get(mruby.ToRuby(mrb, "foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if value.String() != "bar" {
		t.Fatalf("bad: %s", value)
	}

	// Get false type
	value, err = hash.Get(mruby.ToRuby(mrb, "baz"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if valType := value.Type(); valType != mruby.TypeFalse {
		t.Fatalf("bad type: %v", valType)
	}
	if value.String() != "false" {
		t.Fatalf("bad: %s", value)
	}

	// Set
	err = hash.Set(mruby.ToRuby(mrb, "foo"), mruby.ToRuby(mrb, "baz"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	value, err = hash.Get(mruby.ToRuby(mrb, "foo"))
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
	if value.Type() != mruby.TypeArray {
		t.Fatalf("bad: %v", value.Type())
	}
	if value.String() != `["foo", "baz"]` {
		t.Fatalf("bad: %s", value)
	}

	// Delete
	value = hash.Delete(mruby.ToRuby(mrb, "foo"))
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
	value = hash.Delete(mruby.ToRuby(mrb, "nope"))
	if value != nil {
		t.Fatalf("bad: %s", value)
	}
}

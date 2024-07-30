package mruby_test

import (
	"errors"
	"testing"

	mruby "github.com/zhulik/gruby"
)

func testCallback(mrb *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
	return mruby.ToRuby(mrb, 42), nil
}

func testCallbackResult(t *testing.T, value mruby.Value) {
	t.Helper()

	if value.Type() != mruby.TypeFixnum {
		t.Fatalf("bad type: %d", value.Type())
	}

	if mruby.ToGo[int](value) != 42 {
		t.Fatalf("bad: %d", mruby.ToGo[int](value))
	}
}

func testCallbackException(m *mruby.Mrb, self mruby.Value) (mruby.Value, mruby.Value) {
	_, e := m.LoadString(`raise 'Exception'`)
	var err *mruby.ExceptionError
	errors.As(e, &err)
	return nil, err.Value
}

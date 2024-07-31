package gruby_test

import (
	"errors"
	"testing"

	"github.com/zhulik/gruby"
)

func testCallback(mrb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
	return gruby.ToRuby(mrb, 42), nil
}

func testCallbackResult(t *testing.T, value gruby.Value) {
	t.Helper()

	if value.Type() != gruby.TypeFixnum {
		t.Fatalf("bad type: %d", value.Type())
	}

	if gruby.ToGo[int](value) != 42 {
		t.Fatalf("bad: %d", gruby.ToGo[int](value))
	}
}

func testCallbackException(m *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
	_, e := m.LoadString(`raise 'Exception'`)
	var err *gruby.ExceptionError
	errors.As(e, &err)
	return nil, err.Value
}

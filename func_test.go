package mruby

import "testing"

func testCallback(m *Mrb, self Value) (Value, Value) {
	return ToRuby(m, 42), nil
}

func testCallbackResult(t *testing.T, v Value) {
	t.Helper()

	if v.Type() != TypeFixnum {
		t.Fatalf("bad type: %d", v.Type())
	}

	if ToGo[int](v) != 42 {
		t.Fatalf("bad: %d", ToGo[int](v))
	}
}

func testCallbackException(m *Mrb, self Value) (Value, Value) {
	_, e := m.LoadString(`raise 'Exception'`)
	v := e.(*ExceptionError)
	return nil, v.Value
}

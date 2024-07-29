package mruby

import "testing"

func testCallback(m *Mrb, self *MrbValue) (Value, Value) {
	return m.FixnumValue(42), nil
}

func testCallbackResult(t *testing.T, v *MrbValue) {
	t.Helper()

	if v.Type() != TypeFixnum {
		t.Fatalf("bad type: %d", v.Type())
	}

	if v.Fixnum() != 42 {
		t.Fatalf("bad: %d", v.Fixnum())
	}
}

func testCallbackException(m *Mrb, self *MrbValue) (Value, Value) {
	_, e := m.LoadString(`raise 'Exception'`)
	v := e.(*Exception)
	return nil, v.MrbValue
}

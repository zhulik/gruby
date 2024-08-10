package gruby_test

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/zhulik/gruby"
)

func testCallback(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
	return gruby.MustToRuby(grb, 42), nil
}

func testCallbackResult(g G, value gruby.Value) {
	g.Expect(value.Type()).To(gomega.Equal(gruby.TypeFixnum))
	g.Expect(gruby.ToGo[int](value)).To(gomega.Equal(42))
}

func testCallbackException(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
	_, e := grb.LoadString(`raise 'Exception'`)
	var err *gruby.ExceptionError
	errors.As(e, &err)
	return nil, err.Value
}

type G interface {
	Expect(actual any, extra ...any) types.Assertion
}

func NewG(t *testing.T) G {
	t.Helper()

	return gomega.NewWithT(t).
		ConfigureWithFailHandler(func(message string, callerSkip ...int) {
			t.Helper()

			_, file, line, ok := runtime.Caller(callerSkip[0] + 1)
			if !ok {
				t.Fatal("failed to get caller information")
			}
			fmt.Printf("\n%s:%d\n%s\n\n", file, line, message) //nolint:forbidigo
			t.FailNow()
		})
}

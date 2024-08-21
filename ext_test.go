package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestDefaultExtensions(t *testing.T) {
	t.Parallel()

	grb := gruby.Must(gruby.New())
	g := NewG(t)

	val, err := grb.LoadString("gruby_version")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(gruby.ToGo[string](val)).To(HaveLen(5))
}

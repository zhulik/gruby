package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestEnableDisableGC(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	mrb.FullGC()
	mrb.DisableGC()

	_, err := mrb.LoadString("b = []; a = []; a = []")
	g.Expect(err).ToNot(HaveOccurred())

	orig := mrb.LiveObjectCount()
	mrb.FullGC()

	g.Expect(mrb.LiveObjectCount()).To(Equal(orig), "Object count was not what was expected after full GC")

	mrb.EnableGC()
	mrb.FullGC()

	g.Expect(mrb.LiveObjectCount()).To(Equal(orig-1), "Object count was not what was expected after full GC")
}

func TestIsDead(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()

	val, err := mrb.LoadString("$a = []")
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(val.IsDead()).To(BeFalse())

	mrb.Close()

	g.Expect(val.IsDead()).To(BeTrue())
}

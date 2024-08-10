package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestEnableDisableGC(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())
	defer grb.Close()

	grb.FullGC()
	grb.DisableGC()

	_, err := grb.LoadString("b = []; a = []; a = []")
	g.Expect(err).ToNot(HaveOccurred())

	orig := grb.LiveObjectCount()
	grb.FullGC()

	g.Expect(grb.LiveObjectCount()).To(Equal(orig), "Object count was not what was expected after full GC")

	grb.EnableGC()
	grb.FullGC()

	g.Expect(grb.LiveObjectCount()).To(Equal(orig-1), "Object count was not what was expected after full GC")
}

func TestIsDead(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.Must(gruby.New())

	val, err := grb.LoadString("$a = []")
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(val.IsDead()).To(BeFalse())

	grb.Close()

	g.Expect(val.IsDead()).To(BeTrue())
}

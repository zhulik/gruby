package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestArray(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	value, err := mrb.LoadString(`["foo", "bar", "baz", false]`)
	g.Expect(err).ToNot(HaveOccurred())

	array := gruby.ToGo[*gruby.Array](value)

	// Len
	g.Expect(array.Len()).To(Equal(4))

	// Get
	value, err = array.Get(1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("bar"))

	// Get bool
	value, err = array.Get(3)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(value.Type()).To(Equal(gruby.TypeFalse))
	g.Expect(value.String).ToNot(Equal("false"))
}

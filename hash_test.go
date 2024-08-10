package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestHash(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := must(gruby.New())
	defer grb.Close()

	value, err := grb.LoadString(`{"foo" => "bar", "baz" => false}`)
	g.Expect(err).ToNot(HaveOccurred())

	hash := gruby.ToGo[gruby.Hash](value)

	// Get
	value = hash.Get(gruby.MustToRuby(grb, "foo"))
	g.Expect(value.String()).To(Equal("bar"))

	// Get false type
	value = hash.Get(gruby.MustToRuby(grb, "baz"))
	g.Expect(value.Type()).To(Equal(gruby.TypeFalse))
	g.Expect(value.String()).To(Equal("false"))

	// Set
	hash.Set(gruby.MustToRuby(grb, "foo"), gruby.MustToRuby(grb, "baz"))
	value = hash.Get(gruby.MustToRuby(grb, "foo"))
	g.Expect(value.String()).To(Equal("baz"))

	// Keys
	rbKeys := hash.Keys()
	keys := gruby.ToGoArray[string](rbKeys)
	g.Expect(keys).To(Equal([]string{"foo", "baz"}))

	// Delete
	value = hash.Delete(gruby.MustToRuby(grb, "foo"))
	g.Expect(value.String()).To(Equal("baz"))

	rbKeys = hash.Keys()
	keys = gruby.ToGoArray[string](rbKeys)
	g.Expect(keys).To(Equal([]string{"baz"}))

	// Delete non-existing
	value = hash.Delete(gruby.MustToRuby(grb, "nope"))
	g.Expect(value).To(BeNil())
}

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
	value, err = hash.Get(gruby.ToRuby(grb, "foo"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("bar"))

	// Get false type
	value, err = hash.Get(gruby.ToRuby(grb, "baz"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.Type()).To(Equal(gruby.TypeFalse))
	g.Expect(value.String()).To(Equal("false"))

	// Set
	err = hash.Set(gruby.ToRuby(grb, "foo"), gruby.ToRuby(grb, "baz"))
	g.Expect(err).ToNot(HaveOccurred())

	value, err = hash.Get(gruby.ToRuby(grb, "foo"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("baz"))

	// Keys
	rbKeys, err := hash.Keys()
	keys := gruby.ToGoArray[string](rbKeys)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(keys).To(Equal([]string{"foo", "baz"}))

	// Delete
	value = hash.Delete(gruby.ToRuby(grb, "foo"))
	g.Expect(value.String()).To(Equal("baz"))

	rbKeys, err = hash.Keys()
	keys = gruby.ToGoArray[string](rbKeys)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(keys).To(Equal([]string{"baz"}))

	// Delete non-existing
	value = hash.Delete(gruby.ToRuby(grb, "nope"))
	g.Expect(value).To(BeNil())
}

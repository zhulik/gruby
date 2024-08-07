package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestHash(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.New()
	defer mrb.Close()

	value, err := mrb.LoadString(`{"foo" => "bar", "baz" => false}`)
	g.Expect(err).ToNot(HaveOccurred())

	hash := gruby.ToGo[*gruby.Hash](value)

	// Get
	value, err = hash.Get(gruby.ToRuby(mrb, "foo"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("bar"))

	// Get false type
	value, err = hash.Get(gruby.ToRuby(mrb, "baz"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.Type()).To(Equal(gruby.TypeFalse))
	g.Expect(value.String()).To(Equal("false"))

	// Set
	err = hash.Set(gruby.ToRuby(mrb, "foo"), gruby.ToRuby(mrb, "baz"))
	g.Expect(err).ToNot(HaveOccurred())

	value, err = hash.Get(gruby.ToRuby(mrb, "foo"))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal("baz"))

	// Keys
	value, err = hash.Keys()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.Type()).To(Equal(gruby.TypeArray))
	g.Expect(value.String()).To(Equal(`["foo", "baz"]`))

	// Delete
	value = hash.Delete(gruby.ToRuby(mrb, "foo"))
	g.Expect(value.String()).To(Equal("baz"))

	value, err = hash.Keys()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(value.String()).To(Equal(`["baz"]`))

	// Delete non-existing
	value = hash.Delete(gruby.ToRuby(mrb, "nope"))
	g.Expect(value).To(BeNil())
}

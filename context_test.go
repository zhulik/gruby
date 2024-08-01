package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestCompileContextFilename(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	mrb := gruby.NewMrb()
	defer mrb.Close()

	ctx := gruby.NewCompileContext(mrb)
	defer ctx.Close()

	g.Expect(ctx.Filename()).To(BeEmpty())

	ctx.SetFilename("foo")

	g.Expect(ctx.Filename()).To(Equal("foo"))
}

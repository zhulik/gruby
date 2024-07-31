package gruby_test

import (
	"testing"

	"github.com/zhulik/gruby"
)

func TestCompileContextFilename(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	ctx := gruby.NewCompileContext(mrb)
	defer ctx.Close()

	if ctx.Filename() != "" {
		t.Fatalf("bad filename: %s", ctx.Filename())
	}

	ctx.SetFilename("foo")

	if ctx.Filename() != "foo" {
		t.Fatalf("bad filename: %s", ctx.Filename())
	}
}

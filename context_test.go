package mruby_test

import (
	"testing"

	mruby "github.com/zhulik/gruby"
)

func TestCompileContextFilename(t *testing.T) {
	t.Parallel()

	mrb := mruby.NewMrb()
	defer mrb.Close()

	ctx := mruby.NewCompileContext(mrb)
	defer ctx.Close()

	if ctx.Filename() != "" {
		t.Fatalf("bad filename: %s", ctx.Filename())
	}

	ctx.SetFilename("foo")

	if ctx.Filename() != "foo" {
		t.Fatalf("bad filename: %s", ctx.Filename())
	}
}

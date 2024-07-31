package gruby_test

import (
	"testing"

	"github.com/zhulik/gruby"
)

func TestParserGenerateCode(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	parser := gruby.NewParser(mrb)
	defer parser.Close()

	warns, err := parser.Parse(`"foo"`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if warns != nil {
		t.Fatalf("warnings: %v", warns)
	}

	proc := parser.GenerateCode()
	result, err := mrb.Run(proc, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result.String() != "foo" {
		t.Fatalf("bad: %s", result.String())
	}
}

func TestParserParse(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	p := gruby.NewParser(mrb)
	defer p.Close()

	warns, err := p.Parse(`"foo"`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if warns != nil {
		t.Fatalf("warnings: %v", warns)
	}
}

func TestParserParse_error(t *testing.T) {
	t.Parallel()

	mrb := gruby.NewMrb()
	defer mrb.Close()

	p := gruby.NewParser(mrb)
	defer p.Close()

	_, err := p.Parse(`def foo`, nil)
	if err == nil {
		t.Fatal("should have errors")
	}
}

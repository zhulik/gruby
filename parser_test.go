package gruby_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/zhulik/gruby"
)

func TestParserGenerateCode(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.New()
	defer grb.Close()

	parser := gruby.NewParser(grb)
	defer parser.Close()

	warns, err := parser.Parse(`"foo"`, nil)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(warns).To(BeNil())

	proc := parser.GenerateCode()
	result, err := grb.Run(proc, nil)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result.String()).To(Equal("foo"))
}

func TestParserParse(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.New()
	defer grb.Close()

	p := gruby.NewParser(grb)
	defer p.Close()

	warns, err := p.Parse(`"foo"`, nil)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(warns).To(BeNil())
}

func TestParserParse_error(t *testing.T) {
	t.Parallel()
	g := NewG(t)

	grb := gruby.New()
	defer grb.Close()

	p := gruby.NewParser(grb)
	defer p.Close()

	_, err := p.Parse(`def foo`, nil)
	g.Expect(err).To(HaveOccurred())
}

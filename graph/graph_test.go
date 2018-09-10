package graph

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func getReader() (io.Reader, error) {
	buff := bytes.NewBuffer(make([]byte, 0))
	dir, err := os.Getwd()
	if err != nil {
		return buff, fmt.Errorf("unable to get current dir %s", err)
	}
	return os.Open(
		filepath.Join(
			filepath.Dir(dir), "testdata", "so.json",
		),
	)
}

func TestGraph(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	if g.ID() != "so.owl" {
		t.Fatalf("expected Id so.owl does not match %s", g.ID())
	}
	if g.IRI() != "http://purl.obolibrary.org/obo/so.owl" {
		t.Fatalf("expected IRI does not match %s", g.IRI())
	}
	m := g.Meta()
	ver := "http://purl.obolibrary.org/obo/so/so-xp/releases/2015-11-24/so-xp.owl/so.owl"
	if m.Version() != ver {
		t.Fatalf("expected version %s does not match %s", ver, m.Version())
	}
	if len(m.BasicPropertyValues()) != 5 {
		t.Fatalf("expected %d basic properties does not match %d", 5, len(m.BasicPropertyValues()))
	}
	if m.Namespace() != "sequence" {
		t.Fatalf("expected namespace sequence does not match %s", m.Namespace())
	}
}

func TestGraphClassTerm(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	term := "SO_0000340"
	if !g.ExistsTerm(NodeID(term)) {
		t.Fatalf("unable to find term %s", term)
	}
	cht := g.GetTerm(NodeID(term))
	if cht.ID() != NodeID(term) {
		t.Fatalf("did not match term %s with id %s", term, cht.ID())
	}
	if cht.Label() != "chromosome" {
		t.Fatalf("expected label chromosome does not match %s", cht.Label())
	}
	if cht.RdfType() != "CLASS" {
		t.Fatalf("expected type CLASS does not match %s", cht.RdfType())
	}
	if cht.IRI() != "http://purl.obolibrary.org/obo/SO_0000340" {
		t.Fatalf("did not match term %s with iri %s", term, cht.IRI())
	}
	s := cht.Meta().Subsets()
	if len(s) < 1 {
		t.Fatal("expected subset metadata is absent")
	}
	if m, _ := regexp.MatchString("SOFA", s[0]); !m {
		t.Fatalf("expected subset does not match %s", s[0])
	}
	p := cht.Meta().BasicPropertyValues()
	if len(p) < 1 {
		t.Fatal("expected basic propertyvalue metadata is absent")
	}
	if p[0].Value() != "sequence" {
		t.Fatalf("expected basic propertyvalue of sequence does not match %s", p[0].Value())
	}
	if cht.Meta().Namespace() != "sequence" {
		t.Fatalf("expected namespace of sequence does not match %s", cht.Meta().Namespace())
	}
	cm := cht.Meta().Comments()
	if len(cm) < 1 {
		t.Fatal("expected comment is absent")
	}
	if m, _ := regexp.MatchString("MGED", cm[0]); !m {
		t.Fatalf("expected comment does not match %s", cm[0])
	}
}

func TestGraphPropertyTerm(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	if !g.ExistsTerm(NodeID("derives_from")) {
		t.Fatalf("unable to find term %s", "derives_from")
	}
	dft := g.GetTerm(NodeID("derives_from"))
	if dft.ID() != "derives_from" {
		t.Fatalf("did not match term %s with id %s", "derives_from", dft.ID())
	}
	if dft.IRI() != "http://purl.obolibrary.org/obo/so#derives_from" {
		t.Fatalf("did not match term %s with iri %s", "derives_from", dft.IRI())
	}
	s := dft.Meta().Subsets()
	if len(s) < 1 {
		t.Fatal("expected subset metadata is absent")
	}
	if m, _ := regexp.MatchString("SOFA", s[0]); !m {
		t.Fatalf("expected subset does not match %s", s[0])
	}
	p := dft.Meta().BasicPropertyValues()
	if len(p) < 1 {
		t.Fatal("expected basic propertyvalue metadata is absent")
	}
	if p[0].Value() != "sequence" {
		t.Fatalf("expected basic propertyvalue of sequence does not match %s", p[0].Value())
	}
	if dft.Meta().Namespace() != "sequence" {
		t.Fatalf("expected namespace of sequence does not match %s", dft.Meta().Namespace())
	}
}
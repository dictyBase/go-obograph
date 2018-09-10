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

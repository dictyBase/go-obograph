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

func TestGraphProperties(t *testing.T) {
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
	clst := g.TermsByType("CLASS")
	if len(clst) != 2432 {
		t.Fatalf("expected CLASS terms %d does not match %d", 2432, len(clst))
	}
	propt := g.TermsByType("PROPERTY")
	if len(propt) != 82 {
		t.Fatalf("expected PROPERTY terms %d does not match %d", 82, len(propt))
	}
	rels := g.Relationships()
	if len(rels) != 2919 {
		t.Fatalf("expected relationships %d does not match %d", 2919, len(rels))
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

func TestGraphDeprecatedTerm(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	term := "SO_1000100"
	if !g.ExistsTerm(NodeID(term)) {
		t.Fatalf("unable to find term %s", term)
	}
	cht := g.GetTerm(NodeID(term))
	if cht.ID() != NodeID(term) {
		t.Fatalf("did not match term %s with id %s", term, cht.ID())
	}
	if cht.Label() != "mutation_causing_polypeptide_N_terminal_elongation" {
		t.Fatalf("expected label chromosome does not match %s", cht.Label())
	}
	if cht.RdfType() != "CLASS" {
		t.Fatalf("expected type CLASS does not match %s", cht.RdfType())
	}
	if cht.IRI() != "http://purl.obolibrary.org/obo/SO_1000100" {
		t.Fatalf("did not match term %s with iri %s", term, cht.IRI())
	}
	if !cht.IsDeprecated() {
		t.Fatalf("expect term %s to be deprecated", cht.ID())
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

func TestGraphParentTraversal(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	term := "SO_0000336"
	parents := g.Parents(NodeID(term))
	if len(parents) != 2 {
		t.Fatalf("expected %d parents does not match %d", 2, len(parents))
	}
	for _, pterm := range []string{"SO_0000704", "SO_0001411"} {
		if !includesTerm(parents, NodeID(pterm)) {
			t.Fatalf("expected parent term %s does not exist", pterm)
		}
	}
	ancestors := g.Ancestors(NodeID(term))
	if len(ancestors) != 5 {
		t.Fatalf("expected %d ancestors does not match %d", 5, len(ancestors))
	}
	for _, aterm := range []string{"SO_0000704", "SO_0001411", "SO_0005855", "SO_0000001", "SO_0000110"} {
		if !includesTerm(ancestors, NodeID(aterm)) {
			t.Fatalf("expected ancestor term %s does not exist", aterm)
		}
	}
}

func TestGraphChildrenTraversal(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	term := "SO_0001217"
	children := g.Children(NodeID(term))
	if len(children) != 4 {
		t.Fatalf("expected children %d of term %s does not match %d", 4, term, len(children))
	}
	for _, cterm := range []string{"SO_0000548", "SO_0000455", "SO_0000451", "SO_0000693"} {
		if !includesTerm(children, NodeID(cterm)) {
			t.Fatalf("expected child term %s does not exist", cterm)
		}
	}
	desc := g.Descendents(NodeID(term))
	if len(desc) != 9 {
		t.Fatalf("expected %d descendents does not match %d", 9, len(desc))
	}
	for _, dterm := range []string{
		"SO_0000548",
		"SO_0000455",
		"SO_0000451",
		"SO_0000693",
		"SO_0000711",
		"SO_0000712",
		"SO_0000698",
		"SO_0000697",
		"SO_0000710",
	} {
		if !includesTerm(desc, NodeID(dterm)) {
			t.Fatalf("expected child term %s does not exist", dterm)
		}
	}
	descDFS := g.DescendentsDFS(NodeID(term))
	if len(descDFS) != 9 {
		t.Fatalf("expected %d descendents does not match %d", 9, len(descDFS))
	}
	for _, dterm := range []string{
		"SO_0000548",
		"SO_0000455",
		"SO_0000451",
		"SO_0000693",
		"SO_0000711",
		"SO_0000712",
		"SO_0000698",
		"SO_0000697",
		"SO_0000710",
	} {
		if !includesTerm(descDFS, NodeID(dterm)) {
			t.Fatalf("expected child term %s does not exist", dterm)
		}
	}
}

func TestGraphRelationship(t *testing.T) {
	r, err := getReader()
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(r)
	if err != nil {
		t.Fatal(err)
	}
	rel := g.GetRelationship(NodeID("SO_0000704"), NodeID("SO_0001217"))
	if rel.Predicate() != NodeID("is_a") {
		t.Fatalf("expected relationship %s does not match %s", "is_a", rel.Predicate())
	}
	rel2 := g.GetRelationship(NodeID("SO_0000010"), NodeID("SO_0001217"))
	if rel2.Predicate() != NodeID("has_quality") {
		t.Fatalf("expected relationship %s does not match %s", "has_quality", rel2.Predicate())
	}
}

func includesTerm(t []Term, n NodeID) bool {
	for _, v := range t {
		if v.ID() == n {
			return true
		}
	}
	return false
}

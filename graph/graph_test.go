package graph

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	gofn "github.com/repeale/fp-go"
	"github.com/stretchr/testify/require"
)

const (
	SEQ = "sequence"
)

var termPipe = gofn.Map(termToID)

func getReader() (io.Reader, error) {
	buff := bytes.NewBuffer(make([]byte, 0))
	dir, err := os.Getwd()
	if err != nil {
		return buff, fmt.Errorf("unable to get current dir %s", err)
	}
	rdr, err := os.Open(
		filepath.Join(
			filepath.Dir(dir), "testdata", "so.json",
		),
	)
	if err != nil {
		return rdr, fmt.Errorf("error in opening file %s", err)
	}

	return rdr, nil
}

func TestGraphProperties(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	rdr, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(rdr)
	assert.NoError(err, "expect no error from building the graph")
	assert.Equal(grph.ID(), "so.owl", "expect graph Id to match")
	assert.Equal(grph.IRI(), "http://purl.obolibrary.org/obo/so.owl", "expect to match graph IRI")

	mta := grph.Meta()
	ver := "http://purl.obolibrary.org/obo/so/so-xp/releases/2015-11-24/so-xp.owl/so.owl"
	assert.Equal(mta.Version(), ver, "expect to match version")
	assert.Lenf(mta.BasicPropertyValues(), 5, "expected 5 got %d", len(mta.BasicPropertyValues()))
	assert.Equalf(mta.Namespace(), "sequence", "expected sequence namespace got %s", mta.Namespace())
	clst := grph.TermsByType("CLASS")
	assert.Lenf(clst, 2432, "expected 2432 classes got %d", len(clst))
	propt := grph.TermsByType("PROPERTY")
	assert.Lenf(propt, 83, "expected 83 properties got %d", len(propt))
	rels := grph.Relationships()
	assert.Lenf(rels, 2919, "expect 2919 relationships got %d", len(rels))
}

func TestGraphClassTerm(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	rdr, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(rdr)
	assert.NoError(err, "expect no error from building the graph")
	term := "SO_0000340"
	assert.Truef(grph.ExistsTerm(NodeID(term)), "expect to find term %s", term)
	cht := grph.GetTerm(NodeID(term))
	assert.Equalf(cht.ID(), NodeID(term), "expect to match term %s got %s", term, cht.ID())
	assert.Equalf(cht.Label(), "chromosome", "expect to match chromosome got %s", cht.Label())
	assert.Equalf(cht.RdfType(), "CLASS", "expect to match rdf type CLASS got %s", cht.RdfType())
	assert.Equalf(cht.IRI(), "http://purl.obolibrary.org/obo/SO_0000340", "expect to match IRI got %s", cht.IRI())
	assert.Equalf(cht.Meta().Namespace(), "sequence", "expect meta namespace to be sequence got %s", cht.Meta().Namespace())
	sub := cht.Meta().Subsets()
	assert.GreaterOrEqualf(len(sub), 1, "expect 1 or more subsets got %d", len(sub))
	assert.Regexpf("SOFA", sub[0], "expect to match SOFA got %s", sub[0])
	pval := cht.Meta().BasicPropertyValues()
	assert.GreaterOrEqualf(len(pval), 1, "expect 1 or more properties got %d", len(pval))
	assert.Equalf(pval[0].Value(), "sequence", "expect the value to be SEQ got %s", pval[0].Value())
	cmc := cht.Meta().Comments()
	assert.GreaterOrEqualf(len(cmc), 1, "expect 1 or more meta comments got %d", len(cmc))
	assert.Regexpf("MGED", cmc[0], "expect to match MGED got %s", cmc[0])
}

func TestGraphDeprecatedTerm(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	rdr, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(rdr)
	assert.NoError(err, "expect no error from building the graph")

	term := "SO_1000100"
	assert.Truef(grph.ExistsTerm(NodeID(term)), "expect to match %s term", term)
	cht := grph.GetTerm(NodeID(term))
	assert.Equalf(cht.ID(), NodeID(term), "expect to match %s got %s", term, cht.ID())
	assert.Equal(cht.Label(), "mutation_causing_polypeptide_N_terminal_elongation")
	assert.Equalf(cht.RdfType(), "CLASS", "expect to be CLASS but got %s", cht.RdfType())
	assert.Equal(cht.IRI(), "http://purl.obolibrary.org/obo/SO_1000100")
	assert.Truef(cht.IsDeprecated(), "expect %s to be deprecated", cht.ID())
}

func TestGraphPropertyTerm(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	rdr, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(rdr)
	assert.NoError(err, "expect no error from building the graph")

	assert.True(grph.ExistsTerm(NodeID("derives_from")), "expect to find derives_from")
	dft := grph.GetTerm(NodeID("derives_from"))
	assert.Equalf(dft.ID(), NodeID("derives_from"), "expect to match derives_from got %s", dft.ID())
	assert.Equalf(dft.IRI(), "http://purl.obolibrary.org/obo/so#derives_from", "expect to match IRI got %s", dft.IRI())
	assert.Equalf(dft.Meta().Namespace(), "sequence", "expect sequence namespace got %s", dft.Meta().Namespace())
	subs := dft.Meta().Subsets()
	assert.GreaterOrEqualf(len(subs), 1, "expect 1 or more subsets got %d", len(subs))
	assert.Regexpf("SOFA", subs[0], "expect SOFA to match got %s", subs[0])
	props := dft.Meta().BasicPropertyValues()
	assert.GreaterOrEqualf(len(props), 1, "expect 1 or more props got %d", len(props))
	assert.Equalf(props[0].Value(), "sequence", "expect sequence value got %s", props[0].Value())
}

func TestGraphParentTraversal(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	rdr, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(rdr)
	assert.NoError(err, "expect no error from building the graph")

	term := "SO_0000336"
	parents := termPipe(grph.Parents(NodeID(term)))
	assert.Lenf(parents, 2, "expect 2 parents got %d", len(parents))
	for _, pterm := range []string{"SO_0000704", "SO_0001411"} {
		assert.Containsf(parents, NodeID(pterm), "expect %s to be in the parent", pterm)
	}
	ancestors := termPipe(grph.Ancestors(NodeID(term)))
	assert.Lenf(ancestors, 5, "expect 5 ancestors got %d", len(ancestors))
	for _, aterm := range []string{"SO_0000704", "SO_0001411", "SO_0005855", "SO_0000001", "SO_0000110"} {
		assert.Containsf(ancestors, NodeID(aterm), "expect %s to be in the ancestors", aterm)
	}
}

func TestGraphChildrenTraversal(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	rdr, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(rdr)
	assert.NoError(err, "expect no error from building the graph")
	term := "SO_0001217"
	children := termPipe(grph.Children(NodeID(term)))
	assert.Lenf(children, 4, "expect 4 children got %d", len(children))
	for _, cterm := range []string{"SO_0000548", "SO_0000455", "SO_0000451", "SO_0000693"} {
		assert.Containsf(children, NodeID(cterm), "expect %s to be in the children", cterm)
	}
	desc := termPipe(grph.Descendents(NodeID(term)))
	assert.Lenf(desc, 9, "expect 9 descendents got %s", len(desc))
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
		assert.Containsf(desc, NodeID(dterm), "expect %s to be present in descendents", term)
	}
	descDFS := termPipe(grph.DescendentsDFS(NodeID(term)))
	assert.Lenf(descDFS, 9, "expect 9 descendents got %s", len(descDFS))
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
		assert.Containsf(descDFS, NodeID(dterm), "expect %s to be present in descendents", dterm)
	}
}

func TestGraphRelationship(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	r, err := getReader()
	assert.NoError(err, "expect no error from the reader")
	grph, err := BuildGraph(r)
	assert.NoError(err, "expect no error from building the graph")
	rel := grph.GetRelationship(NodeID("SO_0000704"), NodeID("SO_0001217"))
	assert.Equalf(
		rel.Predicate(), NodeID("is_a"),
		"expected relationship is_a got %s", rel.Predicate(),
	)
	rel2 := grph.GetRelationship(NodeID("SO_0000010"), NodeID("SO_0001217"))
	assert.Equalf(
		rel2.Predicate(), NodeID("has_quality"),
		"expected relationship has_quality got %s", rel2.Predicate(),
	)
}

func termToID(trm Term) NodeID {
	return trm.ID()
}

// Package graph provides primitives for building and accessing OBO
// Graphs(graph oriented ontology). The OBO Graphs can be traversed through a
// standard graph oriented API using familiar OBO term and relationship
// concepts.
package graph

import (
	"encoding/json"
	"io"

	"github.com/dictyBase/go-obograph/internal"
	"github.com/dictyBase/go-obograph/model"
	"github.com/dictyBase/go-obograph/schema"
)

// BuildGraph builds an in memory graph from JSON-encoded obograph reader
func BuildGraph(r io.Reader) (OboGraph, error) {
	oj := &schema.OboJSON{}
	err := json.NewDecoder(r).Decode(oj)
	if err != nil {
		return &graph{}, err
	}
	og := oj.Graphs[0]
	g := newOboGraph(
		model.NewMeta(buildGraphMeta(og.Meta)),
		internal.ExtractID(og.ID),
		og.ID,
	)
	// Add the is_a term
	g.AddTerm(buildIsaTerm())
	for _, jn := range og.Nodes {
		g.AddTerm(buildTerm(jn))
	}
	for _, je := range og.Edges {
		err := g.AddRelationshipWithID(
			NodeID(internal.ExtractID(je.Obj)),
			NodeID(internal.ExtractID(je.Sub)),
			NodeID(internal.ExtractID(je.Pred)),
		)
		if err != nil {
			return &graph{}, err
		}
	}
	return g, nil
}

func buildGraphMeta(jm *schema.JSONMeta) *model.MetaOptions {
	m := buildBaseMeta(jm)
	m.Version = jm.Version
	return m
}

func buildTerm(jn *schema.JSONNode) Term {
	return NewTerm(
		NodeID(internal.ExtractID(jn.ID)),
		model.NewMeta(buildTermMeta(jn.Meta)),
		jn.JSONType,
		jn.Lbl,
		jn.ID,
	)
}

func buildIsaTerm() Term {
	return NewTerm(
		NodeID("is_a"),
		model.NewMeta(&model.MetaOptions{}),
		"PROPERTY",
		"is_a",
		"http://www.w3.org/2000/01/rdf-schema#rdfs:subClassOf",
	)
}

func buildTermMeta(jm *schema.JSONMeta) *model.MetaOptions {
	m := buildBaseMeta(jm)
	var syn []*model.Synonym
	if len(jm.Synonyms) > 0 {
		for _, js := range jm.Synonyms {
			if len(js.Xrefs) > 0 {
				syn = append(syn, model.NewSynonymWithRefs(js.Pred, js.Val, js.Xrefs))
			} else {
				syn = append(syn, model.NewSynonym(js.Pred, js.Val))
			}
		}
		m.Synonyms = syn
	}
	m.Definition = model.NewDefinition(
		jm.Definition.Val,
		jm.Definition.Xrefs,
	)
	m.Comments = jm.Comments
	return m
}

func buildBaseMeta(jm *schema.JSONMeta) *model.MetaOptions {
	m := &model.MetaOptions{}
	var p []*model.BasicPropertyValue
	if len(jm.BasicPropertyValues) > 0 {
		for _, bp := range jm.BasicPropertyValues {
			p = append(p, model.NewBasicPropertyValue(bp.Pred, bp.Val))
		}
	}
	m.BaseProps = p
	if len(jm.Subsets) > 0 {
		m.Subsets = jm.Subsets
	}
	if len(jm.Xrefs) > 0 {
		var xref []*model.Xref
		for _, x := range jm.Xrefs {
			xref = append(xref, model.NewXref(x.Val))
		}
		m.Xrefs = xref
	}
	return m
}

// Package graph provides primitives for building and accessing OBO
// Graphs(graph oriented ontology). The OBO Graphs can be traversed through a
// standard graph oriented API using familiar OBO term and relationship
// concepts.
package graph

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/dictyBase/go-obograph/internal"
	"github.com/dictyBase/go-obograph/model"
	"github.com/dictyBase/go-obograph/schema"
)

// BuildGraph builds an in memory graph from JSON-encoded obograph reader.
func BuildGraph(r io.Reader) (OboGraph, error) {
	ojs := &schema.OboJSON{}
	err := json.NewDecoder(r).Decode(ojs)
	if err != nil {
		return &graph{}, fmt.Errorf("error in decoding obograph json %s", err)
	}
	ogf := ojs.Graphs[0]
	grph := newOboGraph(
		model.NewMeta(buildGraphMeta(ogf.Meta)),
		internal.ExtractID(ogf.ID),
		ogf.ID,
	)
	// Add the various owl concepts as obo terms
	grph.AddTerm(buildIsaTerm())
	grph.AddTerm(buildsubPropertyTerm())
	grph.AddTerm(buildinverseOfTerm())
	grph.AddTerm(buildTypeTerm())
	grph.AddTerm(buildtopObjectPropertyTerm())
	for _, jn := range ogf.Nodes {
		grph.AddTerm(buildTerm(jn))
	}
	for _, je := range ogf.Edges {
		err := grph.AddRelationshipWithID(
			NodeID(internal.ExtractID(je.Obj)),
			NodeID(internal.ExtractID(je.Sub)),
			NodeID(internal.ExtractID(je.Pred)),
		)
		if err != nil {
			return &graph{}, fmt.Errorf("error in adding relationship %s", err)
		}
	}

	return grph, nil
}

func buildGraphMeta(jsm *schema.JSONMeta) *model.MetaOptions {
	meta := buildBaseMeta(jsm)
	if len(jsm.Version) > 0 {
		meta.Version = jsm.Version
	}

	return meta
}

func buildTerm(jnn *schema.JSONNode) Term {
	if jnn.Meta != nil {
		return NewTermWithMeta(
			NodeID(internal.ExtractID(jnn.ID)),
			model.NewMeta(buildTermMeta(jnn.Meta)),
			jnn.JSONType,
			jnn.Lbl,
			jnn.ID,
		)
	}

	return NewTerm(
		NodeID(internal.ExtractID(jnn.ID)),
		jnn.JSONType,
		jnn.Lbl,
		jnn.ID,
	)
}

func buildTypeTerm() Term {
	return NewTerm(
		NodeID("type"),
		"PROPERTY",
		"type",
		"https://www.w3.org/1999/02/22-rdf-syntax-ns#type",
	)
}

func buildtopObjectPropertyTerm() Term {
	return NewTerm(
		NodeID("topObjectProperty"),
		"PROPERTY",
		"topObjectProperty",
		"http://www.w3.org/2002/07/owl#topObjectProperty",
	)
}

func buildinverseOfTerm() Term {
	return NewTerm(
		NodeID("inverseOf"),
		"PROPERTY",
		"inverseOf",
		"http://www.w3.org/2002/07/owl#inverseOf",
	)
}

func buildsubPropertyTerm() Term {
	return NewTerm(
		NodeID("subPropertyOf"),
		"PROPERTY",
		"subPropertyOf",
		"http://www.w3.org/2000/01/rdf-schema#subPropertyOf",
	)
}

func buildIsaTerm() Term {
	return NewTerm(
		NodeID("is_a"),
		"PROPERTY",
		"subClassOf",
		"http://www.w3.org/2000/01/rdf-schema#subClassOf",
	)
}

func buildTermMeta(jsm *schema.JSONMeta) *model.MetaOptions {
	meta := buildBaseMeta(jsm)
	if jsm.Synonyms != nil && len(jsm.Synonyms) > 0 {
		var syn []*model.Synonym
		for _, js := range jsm.Synonyms {
			if len(js.Xrefs) > 0 {
				syn = append(
					syn,
					model.NewSynonymWithRefs(js.Pred, js.Val, js.Xrefs),
				)
			} else {
				syn = append(syn, model.NewSynonym(js.Pred, js.Val))
			}
		}
		meta.Synonyms = syn
	}
	if jsm.Definition != nil {
		meta.Definition = model.NewDefinition(
			jsm.Definition.Val,
			jsm.Definition.Xrefs,
		)
	}
	if jsm.Comments != nil && len(jsm.Comments) > 0 {
		meta.Comments = jsm.Comments
	}

	return meta
}

func buildBaseMeta(jsm *schema.JSONMeta) *model.MetaOptions {
	mop := &model.MetaOptions{}
	pval := make([]*model.BasicPropertyValue, 0)
	if jsm.BasicPropertyValues != nil && len(jsm.BasicPropertyValues) > 0 {
		for _, bp := range jsm.BasicPropertyValues {
			pval = append(pval, model.NewBasicPropertyValue(bp.Pred, bp.Val))
		}
		mop.BaseProps = pval
	}
	if jsm.Subsets != nil && len(jsm.Subsets) > 0 {
		mop.Subsets = jsm.Subsets
	}
	if jsm.Xrefs != nil && len(jsm.Xrefs) > 0 {
		var xref []*model.Xref
		for _, x := range jsm.Xrefs {
			xref = append(xref, model.NewXref(x.Val))
		}
		mop.Xrefs = xref
	}
	mop.Deprecated = jsm.Deprecated

	return mop
}

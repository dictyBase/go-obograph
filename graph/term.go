package graph

import (
	"strings"

	"github.com/dictyBase/go-obograph/model"
)

// Term is an interface for obo term(node).
type Term interface {
	// ID is the term's unique identifier
	ID() NodeID
	// HasMeta check for presence of any metadata
	HasMeta() bool
	// Meta returns the term's Meta object
	Meta() *model.Meta
	// RdfType is one defined rdf type, either of CLASS,
	// INDIVIDUAL OR PROPERTY
	RdfType() string
	// Label is a short human readable description of the term
	Label() string
	// IRI represents a stable URL for term's information
	IRI() string
	// IsDeprecated provides the current status of the term
	IsDeprecated() bool
}

type node struct {
	id      NodeID
	meta    *model.Meta
	rdfType string
	lbl     string
	iri     string
}

// NewTerm is the constructor for Term without metadata.
func NewTerm(id NodeID, rdfType, lbl, iri string) Term {
	return &node{
		id:      id,
		rdfType: rdfType,
		lbl:     lbl,
		iri:     iri,
	}
}

// NewTermWithMeta is the constructor for Term with metadata.
func NewTermWithMeta(id NodeID, m *model.Meta, rdfType, lbl, iri string) Term {
	return &node{
		id:      id,
		meta:    m,
		rdfType: rdfType,
		lbl:     lbl,
		iri:     iri,
	}
}

// IsDeprecated provides the current status of the term.
func (n *node) IsDeprecated() bool {
	if !n.HasMeta() {
		return false
	}
	if len(n.meta.BasicPropertyValues()) > 0 {
		for _, p := range n.meta.BasicPropertyValues() {
			if strings.HasSuffix(p.Pred(), "#deprecated") {
				return true
			}
		}
	}

	return n.meta.IsDeprecated()
}

// HasMeta check for presence of any metadata.
func (n *node) HasMeta() bool {
	return n.meta != nil
}

// ID is the term's unique identifier.
func (n *node) ID() NodeID {
	return n.id
}

// Meta returns the term's Meta object.
func (n *node) Meta() *model.Meta {
	if n.meta != nil {
		return n.meta
	}

	return &model.Meta{}
}

// RdfType is one defined rdf type, either of CLASS,
// INDIVIDUAL OR PROPERTY.
func (n *node) RdfType() string {
	return n.rdfType
}

// Label is a short human readable description of the term.
func (n *node) Label() string {
	return n.lbl
}

// IRI represents a stable URL for term's information.
func (n *node) IRI() string {
	return n.iri
}

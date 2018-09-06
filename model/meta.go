// Package model provides various models for structuring extra information
// about terms, relationships and graphs in the OBO Graph.
package model

import (
	"strings"
)

// MetaOptions is a container for various types of metadata
type MetaOptions struct {
	Definition *Definition
	BaseProps  []*BasicPropertyValue
	Synonyms   []*Synonym
	Xrefs      []*Xref
	Comments   []string
	Subsets    []string
	Version    string
}

// Meta is a container for hosting sets of PropertyValue objects
type Meta struct {
	opt *MetaOptions
}

// NewMeta is a constructor for Meta
func NewMeta(opt *MetaOptions) *Meta {
	return &Meta{opt}
}

// BasicPropertyValues are the collection of meta properties
func (m *Meta) BasicPropertyValues() []*BasicPropertyValue {
	return m.opt.BaseProps
}

// Comments are unstructured text information
func (m *Meta) Comments() []string {
	if len(m.opt.Comments) > 0 {
		return m.opt.Comments
	}
	var c []string
	for _, p := range m.BasicPropertyValues() {
		if strings.HasSuffix(p.Pred(), "#comment") {
			c = append(c, p.Value())
		}
	}
	return c
}

// Definition is node definition
func (m *Meta) Definition() *Definition {
	return m.opt.Definition
}

// Synonyms are the synonyms of the nodes
func (m *Meta) Synonyms() []*Synonym {
	return m.opt.Synonyms
}

// Subsets are the subset values of the meta properties
func (m *Meta) Subsets() []string {
	return m.opt.Subsets
}

// Xrefs are slice of all xrefs
func (m *Meta) Xrefs() []*Xref {
	return m.opt.Xrefs
}

// XrefsValues are values of all the xrefs
func (m *Meta) XrefsValues() []string {
	var x []string
	for _, v := range m.Xrefs() {
		x = append(x, v.Value())
	}
	return x
}

// Version the ontology version, will be unset for nodes and edges.
func (m *Meta) Version() string {
	return m.opt.Version
}

// Namespace returns either the default namespace(top level ontology) or the
// individual namespace of a node.
func (m *Meta) Namespace() string {
	for _, p := range m.BasicPropertyValues() {
		if strings.HasSuffix(p.Pred(), "#hasOBONamespace") {
			return p.Value()
		}
		if strings.HasSuffix(p.Pred(), "#default-namespace") {
			return p.Value()
		}
	}
	return ""

}

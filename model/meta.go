// Package model provides various models for structuring extra information
// about terms, relationships and graphs in the OBO Graph.
package model

import (
	"strings"
)

// MetaOptions is a container for various types of metadata.
type MetaOptions struct {
	Definition *Definition
	BaseProps  []*BasicPropertyValue
	Synonyms   []*Synonym
	Xrefs      []*Xref
	Comments   []string
	Subsets    []string
	Version    string
	Deprecated bool
}

// Meta is a container for hosting sets of PropertyValue objects.
type Meta struct {
	opt *MetaOptions
}

// NewMeta is a constructor for Meta.
func NewMeta(opt *MetaOptions) *Meta {
	return &Meta{opt}
}

// BasicPropertyValues are the collection of meta properties.
func (m *Meta) BasicPropertyValues() []*BasicPropertyValue {
	var b []*BasicPropertyValue
	if len(m.opt.BaseProps) > 0 {
		return m.opt.BaseProps
	}

	return b
}

// Comments are unstructured text information.
func (m *Meta) Comments() []string {
	if len(m.opt.Comments) > 0 {
		return m.opt.Comments
	}
	var comm []string
	for _, p := range m.BasicPropertyValues() {
		if strings.HasSuffix(p.Pred(), "#comment") {
			comm = append(comm, p.Value())
		}
	}

	return comm
}

// Definition is node definition.
func (m *Meta) Definition() *Definition {
	var d *Definition
	if m.opt.Definition != nil {
		return m.opt.Definition
	}

	return d
}

// Synonyms are the synonyms of the nodes.
func (m *Meta) Synonyms() []*Synonym {
	var s []*Synonym
	if len(m.opt.Synonyms) > 0 {
		return m.opt.Synonyms
	}

	return s
}

// Subsets are the subset values of the meta properties.
func (m *Meta) Subsets() []string {
	var s []string
	if len(m.opt.Subsets) > 0 {
		return m.opt.Subsets
	}

	return s
}

// Xrefs are slice of all xrefs.
func (m *Meta) Xrefs() []*Xref {
	var x []*Xref
	if len(m.opt.Xrefs) > 0 {
		return m.opt.Xrefs
	}

	return x
}

// XrefsValues are values of all the xrefs.
func (m *Meta) XrefsValues() []string {
	x := make([]string, 0)
	for _, v := range m.Xrefs() {
		x = append(x, v.Value())
	}

	return x
}

// Version the ontology version, will be unset for nodes and edges.
func (m *Meta) Version() string {
	var v string
	if len(m.opt.Version) > 0 {
		return m.opt.Version
	}

	return v
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

// IsDeprecated returns a boolean indicating whether the meta information is deprecated.
func (m *Meta) IsDeprecated() bool {
	return m.opt.Deprecated
}

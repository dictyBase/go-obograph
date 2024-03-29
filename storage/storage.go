// Package storage provides type definition for managing OBO
// graphs in a persistent storage
package storage

import (
	"github.com/dictyBase/go-obograph/graph"
)

// Stats provides statistics about terms.
type Stats struct {
	Created int
	Updated int
	Deleted int
}

// DataSource represents interface for storing and retrieving
// OBO graphs.
type DataSource interface {
	// SaveOboGraphInfo perist OBO graphs metadata in the storage
	SaveOboGraphInfo(graph.OboGraph) error
	// UpdateOboGraphInfo update OBO graph metadata in the storage
	UpdateOboGraphInfo(graph.OboGraph) error
	// ExistOboGraph checks for existence of a particular OBO graph
	ExistsOboGraph(graph.OboGraph) bool
	// SaveTerms persist all terms in the storage
	SaveTerms(graph.OboGraph) (int, error)
	// UpdateTerms update existing terms in the storage
	UpdateTerms(graph.OboGraph) (int, error)
	// SaveorUpdateTerms either insert and update terms in the storage
	SaveOrUpdateTerms(graph.OboGraph) (*Stats, error)
	// SaveRelationships persist all relationships in the storage
	SaveRelationships(graph.OboGraph) (int, error)
	// SaveNewRelationships skips the existing one and saves only the new relationships
	SaveNewRelationships(graph.OboGraph) (int, error)
}

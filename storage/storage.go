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
	SaveOboGraphInfo(grph graph.OboGraph) error
	// UpdateOboGraphInfo update OBO graph metadata in the storage
	UpdateOboGraphInfo(grph graph.OboGraph) error
	// ExistOboGraph checks for existence of a particular OBO graph
	ExistsOboGraph(grph graph.OboGraph) bool
	// SaveTerms persist all terms in the storage
	SaveTerms(grph graph.OboGraph) (int, error)
	// UpdateTerms update existing terms in the storage
	UpdateTerms(grph graph.OboGraph) (int, error)
	// SaveorUpdateTerms either insert and update terms in the storage
	SaveOrUpdateTerms(grph graph.OboGraph) (*Stats, error)
	// SaveRelationships persist all relationships in the storage
	SaveRelationships(grph graph.OboGraph) (int, error)
	// SaveNewRelationships skips the existing one and saves only the new relationships
	SaveNewRelationships(grph graph.OboGraph) (int, error)
}

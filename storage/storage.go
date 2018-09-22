// Package storage provides type definition for managing OBO
// graphs in a persistent storage
package storage

import (
	"github.com/dictyBase/go-obograph/graph"
)

// DataSource represents interface for storing and retrieving
// OBO graphs
type DataSource interface {
	// SaveOboGraphInfo perist OBO graphs metadata in the storage
	SaveOboGraphInfo(graph.OboGraph) error
	// ExistOboGraph checks for existence of a particular OBO graph
	ExistsOboGraph(graph.OboGraph) bool
	// SaveTerms persist slice of terms in the storage
	SaveTerms(graph.OboGraph) (int, error)
	// UpdateTerms update slice of terms in the storage
	UpdateTerms(graph.OboGraph) (int, error)
	// SaveorUpdateTerms either insert or update a slice of terms
	SaveOrUpdateTerms(graph.OboGraph) (int, error)
	// SaveRelationships persist slice of relationships in the storage
	SaveRelationships(graph.OboGraph) (int, error)
	// SaveNewRelationships skips the existing one and saves only the new relationships
	SaveNewRelationships(graph.OboGraph) (int, error)
}

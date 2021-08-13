package arangodb

import (
	"context"

	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
)

type OntoCollection struct {
	Term driver.Collection
	Rel  driver.Collection
	Cv   driver.Collection
	Obog driver.Graph
}

// CreateCollection creates all the necessary collections, graph and index required
// for persisting obojson ontology in arangodb
func CreateCollection(db *manager.Database, collP *CollectionParams) (*OntoCollection, error) {
	oc, err := docCollection(db, collP)
	if err != nil {
		return oc, err
	}
	obog, err := db.FindOrCreateGraph(
		collP.OboGraph,
		[]driver.EdgeDefinition{{
			Collection: oc.Rel.Name(),
			From:       []string{oc.Term.Name()},
			To:         []string{oc.Term.Name()},
		}})
	if err != nil {
		return oc, err
	}
	oc.Obog = obog
	_, _, err = oc.Term.EnsurePersistentIndex(
		context.Background(),
		[]string{"label"},
		&driver.EnsurePersistentIndexOptions{
			Name:         "label-idx",
			InBackground: true,
		})
	return oc, err
}

func docCollection(db *manager.Database, collP *CollectionParams) (*OntoCollection, error) {
	oc := &OntoCollection{}
	termc, err := db.FindOrCreateCollection(
		collP.Term,
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return oc, err
	}
	relc, err := db.FindOrCreateCollection(
		collP.Relationship,
		&driver.CreateCollectionOptions{Type: driver.CollectionTypeEdge},
	)
	if err != nil {
		return oc, err
	}
	graphc, err := db.FindOrCreateCollection(
		collP.GraphInfo,
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return oc, err
	}
	oc.Term = termc
	oc.Rel = relc
	oc.Cv = graphc
	return oc, nil
}

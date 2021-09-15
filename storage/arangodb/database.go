package arangodb

import (
	"context"

	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
)

type OntoCollection struct {
	Term  driver.Collection
	Rel   driver.Collection
	Cv    driver.Collection
	Obog  driver.Graph
	db    *manager.Database
	collP *CollectionParams
}

func (oc *OntoCollection) docCollection() error {
	db := oc.db
	collP := oc.collP
	termc, err := db.FindOrCreateCollection(
		collP.Term,
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return err
	}
	relc, err := db.FindOrCreateCollection(
		collP.Relationship,
		&driver.CreateCollectionOptions{Type: driver.CollectionTypeEdge},
	)
	if err != nil {
		return err
	}
	graphc, err := db.FindOrCreateCollection(
		collP.GraphInfo,
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return err
	}
	oc.Term = termc
	oc.Rel = relc
	oc.Cv = graphc
	return nil
}

func (oc *OntoCollection) graphAndIndex() error {
	db := oc.db
	collP := oc.collP
	obog, err := db.FindOrCreateGraph(
		collP.OboGraph,
		[]driver.EdgeDefinition{{
			Collection: oc.Rel.Name(),
			From:       []string{oc.Term.Name()},
			To:         []string{oc.Term.Name()},
		}})
	if err != nil {
		return err
	}
	_, _, err = oc.Term.EnsurePersistentIndex(
		context.Background(),
		[]string{"label"},
		&driver.EnsurePersistentIndexOptions{
			Name:         "label-idx",
			InBackground: true,
		})
	oc.Obog = obog
	return nil
}

// CreateCollection creates all the necessary collections, graph and index required
// for persisting obojson ontology in arangodb
func CreateCollection(db *manager.Database, collP *CollectionParams) (*OntoCollection, error) {
	oc := &OntoCollection{db: db, collP: collP}
	if err := oc.docCollection(); err != nil {
		return oc, err
	}
	if err := oc.graphAndIndex(); err != nil {
		return oc, err
	}
	return oc, nil
}

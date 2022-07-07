package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
)

type OntoCollection struct {
	Term  driver.Collection
	Rel   driver.Collection
	Cv    driver.Collection
	Obog  driver.Graph
	dbh   *manager.Database
	collP *CollectionParams
}

func (oc *OntoCollection) docCollection() error {
	dbh := oc.dbh
	collP := oc.collP
	termc, err := dbh.FindOrCreateCollection(
		collP.Term,
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return err
	}
	relc, err := dbh.FindOrCreateCollection(
		collP.Relationship,
		&driver.CreateCollectionOptions{Type: driver.CollectionTypeEdge},
	)
	if err != nil {
		return err
	}
	graphc, err := dbh.FindOrCreateCollection(
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
	dbh := oc.dbh
	collP := oc.collP
	obog, err := dbh.FindOrCreateGraph(
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
	if err != nil {
		return fmt.Errorf("error in creating index %s", err)
	}
	oc.Obog = obog

	return nil
}

// CreateCollection creates all the necessary collections, graph and index required
// for persisting obojson ontology in arangodb.
func CreateCollection(dbh *manager.Database, collP *CollectionParams) (*OntoCollection, error) {
	ocn := &OntoCollection{dbh: dbh, collP: collP}
	if err := ocn.docCollection(); err != nil {
		return ocn, err
	}
	if err := ocn.graphAndIndex(); err != nil {
		return ocn, err
	}

	return ocn, nil
}

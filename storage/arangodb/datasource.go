package arangodb

import (
	"context"
	"time"

	"gopkg.in/go-playground/validator.v9"

	driver "github.com/arangodb/go-driver"
	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
	"github.com/dictyBase/go-obograph/storage/arangodb/manager"
)

// ConnectParams are the parameters required for connecting to arangodb
type ConnectParams struct {
	User     string `validate:"required"`
	Pass     string `validate:"required"`
	Database string `validate:"required"`
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	Istls    bool
}

// CollectionParams are the arangodb collections required for storing
// OBO graphs
type CollectionParams struct {
	// Term is the collection for storing term(nodes)
	Term string `validate:"required"`
	// Relationship is the collection for storing relationship(edges)
	Relationship string `validate:"required"`
	// GraphInfo is the collection for storing graph metadata
	GraphInfo string `validate:"required"`
}

func NewDataSource(connP *ConnectParams, collP *CollectionParams) (storage.DataSource, error) {
	var ds *arangoSource
	validate := validator.New()
	if err := validate.Struct(connP); err != nil {
		return ds, err
	}
	if err := validate.Struct(collP); err != nil {
		return ds, err
	}
	sess, err := manager.Connect(
		connP.Host,
		connP.User,
		connP.Pass,
		connP.Port,
		connP.Istls,
	)
	if err != nil {
		return ds, err
	}
	db, err := sess.DB(connP.Database)
	if err != nil {
		return ds, err
	}
	termc, err := db.Collection(collP.Term)
	if err != nil {
		return ds, err
	}
	relc, err := db.Collection(coll.Relationship)
	if err != nil {
		return ds, err
	}
	graphc, err := db.Collection(coll.GraphInfo)
	if err != nil {
		return ds, err
	}
	return &arangoSource{
		sess:     sess,
		database: db,
		termc:    termc,
		relc:     relc,
		graphc:   graphc,
	}, nil
}

type arangoSource struct {
	sess     *manager.Session
	database *manager.Database
	termc    driver.Collection
	relc     driver.Collection
	graphc   driver.Collection
}

func (a *arangoSource) SaveOboGraphInfo(g graph.OboGraph) error {
	var dp []*dbGraphProps
	for _, p := range g.Meta().BasicPropertyValues() {
		dp = append(dp, &dbGraphProps{
			pred:  p.Pred(),
			value: p.Value(),
			curie: curieMap[p.Pred()],
		})
	}
	dg := &dbGraphInfo{
		id:        g.ID(),
		iri:       g.IRI(),
		label:     g.Label(),
		createdAt: g.Timestamp(),
		updatedAt: g.Timestamp(),
		metadata: &dbGraphMeta{
			namespace:  g.Meta().Namespace(),
			version:    g.Meta().Version(),
			properties: dp,
		},
	}
	ctx := driver.WithSilent(context.Background())
	_, err := a.graphc.CreateDocument(ctx, dg)
	return err
}

func (a *arangoSource) ExistsOboGraph(g graph.OboGraph) bool {
	ctx := driver.WithQueryCount(context.Background())
	query := `FOR d in @@collection
				FILTER d.id == @identifier
				RETURN d`
	bindVars := map[string]interface{}{
		"@collection": a.graphc.Name(),
		"identifier":  g.ID(),
	}
	cursor, err := a.database.Query(ctx, query, bindVars)
	if err != nil {
		return false
	}
	defer cursor.Close()
	if cursor.Count() > 0 {
		return true
	}
	return false
}

func (a *arangoSource) IsUpdatedOboGraph(g graph.OboGraph) bool {
	if !a.ExistsOboGraph(g) {
		return true
	}
	var ts time.Time
	query := `FOR d in @@collection
				FILTER d.identifier == @identifier
				LIMIT 1
				RETURN d.updated_at `
	bindVars := map[string]interface{}{
		"@collection": a.graphc.Name(),
		"identifier":  g.ID(),
	}
	ctx := driver.WithQueryCache(context.Background(), true)
	cursor, err := a.database.Query(ctx, query, bindVars)
	defer cursor.Close()
	_, err := cursor.ReadDocument(ctx, ts)
	if err != nil {
		return err
	}
	return g.Timestamp().After(ts)
}

func (a *arangoSource) SaveTerms(ts []graph.Term) (int, error) {
	stat, err := a.termc.ImportDocuments(
		context.Background(),
		todbTerm(ts),
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return 0, err
	}
	return int(stat.Created), nil
}

func (a *arangoSource) UpdateTerms(ts []graph.Term) (int, error) {
	panic("not implemented")
}

func (a *arangoSource) SaveOrUpdateTerms(ts []graph.Term) (int, error) {
	panic("not implemented")
}

func (a *arangoSource) SaveRelationships(rs []graph.Relationship) (int, error) {
	panic("not implemented")
}

func (a *arangoSource) SaveNewRelationships(rs []graph.Relationship) (int, error) {
	panic("not implemented")
}

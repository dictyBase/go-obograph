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
	query := manager.NewAqlStruct().
		For("d", a.graphc.Name()).
		Filter("d", Fil("id", "eq", g.ID())).
		Return("d")
	count, err := a.database.Count(query.Generate())
	if err != nil {
		return false
	}
	if count > 0 {
		return true
	}
	return false
}

func (a *arangoSource) IsUpdatedOboGraph(g graph.OboGraph) bool {
	if !a.ExistsOboGraph(g) {
		return true
	}
	query := manager.NewAqlStruct().
		For("d", a.graphc.Name()).
		Filter("d", Fil("id", "eq", g.ID())).
		Limit(1).
		Return("d.updated_at")
	res, err := a.database.Get(query.Generate())
	var s string
	if err := res.Read(&s); err != nil {
		return false
	}
	ts, err := time.Parse("02:01:2006 15:04", s)
	if err != nil {
		return false
	}
	return g.Timestamp().After(ts)
}

func (a *arangoSource) SaveTerms(terms []graph.Term) (int, error) {
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

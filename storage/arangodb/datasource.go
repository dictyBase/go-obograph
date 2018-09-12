package arangodb

import (
	"context"
	"crypto/tls"
	"fmt"

	"gopkg.in/go-playground/validator.v9"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
)

// ConnectParams are the parameters required for connecting to arangodb
type ConnectParams struct {
	User     string `validate:"required"`
	Pass     string `validate:"required"`
	Database string `validate:"required"`
	Host     string `validate:"required"`
	Port     string `validate:"required"`
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
	connConf := http.ConnectionConfig{
		Endpoints: []string{
			fmt.Sprintf("http://%s:%s", connP.Host, connP.Port),
		},
	}
	if connP.Istls {
		connConf.Endpoints = []string{
			fmt.Sprintf("https://%s:%s", connP.Host, connP.Port),
		}
		connConf.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	conn, err := http.NewConnection(connConf)
	if err != nil {
		return ds, fmt.Errorf("could not connect %s", err)
	}
	client, err := driver.NewClient(
		driver.ClientConfig{
			Connection: conn,
			Authentication: driver.BasicAuthentication(
				connP.User,
				connP.Pass,
			),
		})
	if err != nil {
		return ds, fmt.Errorf("could not get a client instance %s", err)
	}
	db, err := getDatabase(connP.Database, client)
	if err != nil {
		return ds, err
	}
	termc, err := getCollection(db, collP.Term)
	if err != nil {
		return ds, err
	}
	relc, err := getCollection(db, coll.Relationship)
	if err != nil {
		return ds, err
	}
	graphc, err := getCollection(db, coll.GraphInfo)
	if err != nil {
		return ds, err
	}
	return &arangoSource{
		database: db,
		termc:    termc,
		relc:     relc,
		graphc:   graphc,
	}, nil
}

type arangoSource struct {
	database driver.Database
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
		UpdatedAt: g.Timestamp(),
		metadata: &dbGraphMeta{
			namespace:  g.Meta().Namespace(),
			version:    g.Meta().Version(),
			properties: dp,
		},
	}
	_, err := a.graphc.CreateDocument(context.Background(), dg)
	return err
}

func (a *arangoSource) ExistsOboGraph(g graph.OboGraph) bool {
	panic("not implemented")
}

func (a *arangoSource) IsUpdatedOboGraph(g graph.OboGraph) bool {
	panic("not implemented")
}

func (a *arangoSource) SaveTerms(ts []graph.Term) (int, error) {
	panic("not implemented")
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

package arangodb

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/go-playground/validator.v9"

	driver "github.com/arangodb/go-driver"
	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
	"github.com/dictyBase/go-obograph/storage/arangodb/manager"
)

var sMap map[graph.NodeID]string = make(map[graph.NodeID]string)
var pMap map[graph.NodeID]string = make(map[graph.NodeID]string)
var oMap map[graph.NodeID]string = make(map[graph.NodeID]string)

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
	var dbterms []*dbTerm
	for _, t := range terms {
		dbterms = append(dbterms, a.todbTerm(t))
	}
	stat, err := a.termc.ImportDocuments(
		context.Background(),
		dbterms,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return 0, err
	}
	return int(stat.Created), nil
}

func (a *arangoSource) SaveRelationships(rels []graph.Relationship) (int, error) {
	var dbrs []*db.Relationship
	for _, r := range rels {
		dbrel, err := a.todbRelationhip(r)
		if err != nil {
			return 0, err
		}
		dbrs = append(dbrs, dbrel)
	}
	stat, err := a.relc.ImportDocuments(
		context.Background(),
		dbrs,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return 0, err
	}
	return int(stat.Created), nil
}

func (a *arangoSource) UpdateTerms(terms []graph.Term) (int, error) {
	panic("not implemented")
}

func (a *arangoSource) SaveOrUpdateTerms(terms []graph.Term) (int, error) {
	panic("not implemented")
}

func (a *arangoSource) SaveNewRelationships(rels []graph.Relationship) (int, error) {
	panic("not implemented")
}

func (a *arangoSource) todbTerm(t graph.Term) *dbTerm {
	var dbm *dbTermMeta
	var dps []*dbGraphProps
	for _, p := range t.Meta().BasicPropertyValues() {
		dps = append(dps, &dbGraphProps{
			pred:  p.Pred(),
			value: p.Value(),
			curie: curieMap[p.Pred()],
		})
	}
	dbm.properties = dps

	if len(t.Meta().Xrefs()) > 0 {
		var dbx []*dbMetaXref
		for _, r := range t.Meta().Xrefs() {
			dbx = append(dbx, &dbMetaXref{value: r.Value()})
		}
		dbm.xrefs = dbx
	}

	if len(t.Meta().Synonyms()) > 0 {
		var dbs []*dbMetaSynonym
		for _, s := range t.Meta().Synonyms() {
			dbs = append(dbs, &dbMetaSynonym{
				value:   s.Value(),
				pred:    s.Pred(),
				scope:   s.Scope(),
				isExact: s.IsExact(),
				xrefs:   s.Xrefs(),
			})
		}
	}
	dbm.synonyms = dbs

	if len(t.Meta().Comments()) > 1 {
		dbm.comments = t.Meta().Comments()
	}
	if len(t.Meta().Subsets()) > 1 {
		dbm.subsets = t.Meta().Subsets()
	}
	if t.Meta().Definition() != nil {
		dbm.definition = &dbMetaDefinition{
			value: t.Meta().Definition().Value(),
			xrefs: t.Meta().Definition().Xrefs(),
		}
	}
	dbm.namespace = t.Meta().Namespace()
	return &dbTerm{
		id:       t.ID(),
		iri:      t.IRI(),
		label:    t.Label(),
		rdfType:  t.RdfType(),
		metadata: dbm,
	}
}

func (a *arangoSource) todbRelationhip(r graph.Relationship) (*db.Relationship, error) {
	dbr := &dbRelationship{}
	if v, ok := oMap[r.Object()]; ok {
		dbr.from = v
	} else {
		id, err := a.getDocId(r.Object())
		if err != nil {
			return dbr, err
		}
		oMap[r.Object()] = id
		dbr.from = id
	}
	if v, ok := oMap[r.Subject()]; ok {
		dbr.to = v
	} else {
		id, err := a.getDocId(r.Subject())
		if err != nil {
			return dbr, err
		}
		oMap[r.Subject()] = id
		dbr.from = id
	}
	if v, ok := oMap[r.Predicate()]; ok {
		dbr.predicate = v
	} else {
		id, err := a.getDocId(r.Predicate())
		if err != nil {
			return dbr, err
		}
		oMap[r.Predicate()] = id
		dbr.predicate = id
	}
	return dbr, nil
}

func (a *arangoSource) getDocId(nid graph.NodeID) (string, error) {
	var id string
	query := manager.NewAqlStruct().
		For("d", a.termc.Name()).
		Filter("d", Fil("id", "eq", string(nid))).
		Return("d._id")
	res, err := a.database.Get(query.Generate())
	if err != nil {
		return id, err
	}
	if res.IsEmpty() {
		return id, fmt.Errorf("object %s is absent in database", nid)
	}
	err := res.Read(&id)
	return id, err
}

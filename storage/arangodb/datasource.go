package arangodb

import (
	"context"
	"fmt"

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
	relc, err := db.Collection(collP.Relationship)
	if err != nil {
		return ds, err
	}
	graphc, err := db.Collection(collP.GraphInfo)
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

// SaveOboGraphInfo perist OBO graphs metadata in the storage
func (a *arangoSource) SaveOboGraphInfo(g graph.OboGraph) error {
	dg := dbGraphInfo{
		Id:       g.ID(),
		IRI:      g.IRI(),
		Label:    g.Label(),
		Metadata: a.todbGraphMeta(g),
	}
	ctx := driver.WithSilent(context.Background())
	_, err := a.graphc.CreateDocument(ctx, dg)
	return err
}

// ExistOboGraph checks for existence of a particular OBO graph
func (a *arangoSource) ExistsOboGraph(g graph.OboGraph) bool {
	query := manager.NewAqlStruct().
		For("d", a.graphc.Name()).
		Filter("d", manager.Fil("id", "eq", g.ID()), true).
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

func (a *arangoSource) SaveTerms(g graph.OboGraph) (int, error) {
	id, err := a.graphDocId(g)
	if err != nil {
		return 0, err
	}
	var dbterms []*dbTerm
	for _, t := range g.Terms() {
		dbterms = append(dbterms, a.todbTerm(id, t))
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

func (a *arangoSource) SaveRelationships(g graph.OboGraph) (int, error) {
	var dbrs []*dbRelationship
	for _, r := range g.Relationships() {
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

func (a *arangoSource) UpdateTerms(g graph.OboGraph) (int, error) {
	return 0, nil
}

func (a *arangoSource) SaveOrUpdateTerms(g graph.OboGraph) (int, int, error) {
	key, err := a.graphDocKey(g)
	if err != nil {
		return 0, 0, err
	}
	dg := dbGraphInfo{
		Metadata: a.todbGraphMeta(g),
	}
	base := manager.NewAqlStruct().
		For("d", a.graphc.Name()).
		Filter("d", manager.Fil("id", "eq", g.ID()), true)
	query := fmt.Sprintf(`
		%s UPDATE d WITH @data IN %s`,
		base.Generate(), a.graphc.Name(),
	)
	bindVars := map[string]interface{}{
		"data": dg,
	}
	err := a.database.Do(query, bindVars)
	if err != nil {
		return 0, 0, fmt.Errorf("error in updating the graph %s", err)
	}
	return 1, 1, nil
}

func (a *arangoSource) SaveNewRelationships(g graph.OboGraph) (int, error) {
	return 0, nil
}

func (a *arangoSource) todbGraphMeta(g graph.OboGraph) *dbGraphMeta {
	var dp []*dbGraphProps
	for _, p := range g.Meta().BasicPropertyValues() {
		dp = append(dp, &dbGraphProps{
			Pred:  p.Pred(),
			Value: p.Value(),
			Curie: curieMap[p.Pred()],
		})
	}
	return &dbGraphMeta{
		Namespace:  g.Meta().Namespace(),
		Version:    g.Meta().Version(),
		Properties: dp,
	}
}

func (a *arangoSource) todbTerm(id string, t graph.Term) *dbTerm {
	dbt := &dbTerm{
		Id:         string(t.ID()),
		Iri:        t.IRI(),
		Label:      t.Label(),
		RdfType:    t.RdfType(),
		Deprecated: t.IsDeprecated(),
		GraphId:    id,
	}
	if !t.HasMeta() {
		return dbt
	}

	dbm := new(dbTermMeta)
	var dps []*dbGraphProps
	if len(t.Meta().BasicPropertyValues()) > 0 {
		for _, p := range t.Meta().BasicPropertyValues() {
			dps = append(dps, &dbGraphProps{
				Pred:  p.Pred(),
				Value: p.Value(),
				Curie: curieMap[p.Pred()],
			})
		}
		dbm.Properties = dps
	}

	if len(t.Meta().Xrefs()) > 0 {
		var dbx []*dbMetaXref
		for _, r := range t.Meta().Xrefs() {
			dbx = append(dbx, &dbMetaXref{Value: r.Value()})
		}
		dbm.Xrefs = dbx
	}

	if len(t.Meta().Synonyms()) > 0 {
		var dbs []*dbMetaSynonym
		for _, s := range t.Meta().Synonyms() {
			dbs = append(dbs, &dbMetaSynonym{
				Value:   s.Value(),
				Pred:    s.Pred(),
				Scope:   s.Scope(),
				IsExact: s.IsExact(),
				Xrefs:   s.Xrefs(),
			})
		}
		dbm.Synonyms = dbs
	}

	if len(t.Meta().Comments()) > 0 {
		dbm.Comments = t.Meta().Comments()
	}
	if len(t.Meta().Subsets()) > 0 {
		dbm.Subsets = t.Meta().Subsets()
	}
	if t.Meta().Definition() != nil {
		dbm.Definition = &dbMetaDefinition{
			Value: t.Meta().Definition().Value(),
			Xrefs: t.Meta().Definition().Xrefs(),
		}
	}
	dbm.Namespace = t.Meta().Namespace()
	dbt.Metadata = dbm
	return dbt
}

func (a *arangoSource) todbRelationhip(r graph.Relationship) (*dbRelationship, error) {
	dbr := &dbRelationship{}
	if v, ok := oMap[r.Object()]; ok {
		dbr.From = v
	} else {
		id, err := a.getDocId(r.Object())
		if err != nil {
			return dbr, err
		}
		oMap[r.Object()] = id
		dbr.From = id
	}
	if v, ok := oMap[r.Subject()]; ok {
		dbr.To = v
	} else {
		id, err := a.getDocId(r.Subject())
		if err != nil {
			return dbr, err
		}
		oMap[r.Subject()] = id
		dbr.To = id
	}
	if v, ok := oMap[r.Predicate()]; ok {
		dbr.Predicate = v
	} else {
		id, err := a.getDocId(r.Predicate())
		if err != nil {
			return dbr, err
		}
		oMap[r.Predicate()] = id
		dbr.Predicate = id
	}
	return dbr, nil
}

func (a *arangoSource) getDocId(nid graph.NodeID) (string, error) {
	var id string
	query := manager.NewAqlStruct().
		For("d", a.termc.Name()).
		Filter("d", manager.Fil("id", "eq", string(nid)), true).
		Return("d._id")
	res, err := a.database.Get(query.Generate())
	if err != nil {
		return id, err
	}
	if res.IsEmpty() {
		return id, fmt.Errorf("object %s is absent in database", nid)
	}
	err = res.Read(&id)
	return id, err
}

func (a *arangoSource) graphDocId(g graph.OboGraph) (string, error) {
	query := manager.NewAqlStruct().
		For("d", a.graphc.Name()).
		Filter("d", manager.Fil("id", "eq", g.ID()), true).
		Return("d._id")
	res, err := a.database.Get(query.Generate())
	if err != nil {
		return 0, err
	}
	if res.IsEmpty() {
		return 0, fmt.Errorf("graph id %s is absent from database", g.ID())
	}
	var id string
	err = res.Read(&id)
	if err != nil {
		return 0, err
	}
	return id, err
}

func (a *arangoSource) graphDocKey(g graph.OboGraph) (string, error) {
	query := manager.NewAqlStruct().
		For("d", a.graphc.Name()).
		Filter("d", manager.Fil("id", "eq", g.ID()), true).
		Return("d._key")
	res, err := a.database.Get(query.Generate())
	if err != nil {
		return 0, err
	}
	if res.IsEmpty() {
		return 0, fmt.Errorf("graph id %s is absent from database", g.ID())
	}
	var key string
	err = res.Read(&key)
	if err != nil {
		return 0, err
	}
	return key, err
}

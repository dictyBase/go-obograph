package arangodb

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"

	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/go-obograph/generate"
	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
)

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
	// OboGraph is the named graph for connecting term and relationship collections
	OboGraph string `validate:"required"`
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
	sess, db, err := manager.NewSessionDb(&manager.ConnectParams{
		User:     connP.User,
		Pass:     connP.Pass,
		Database: connP.Database,
		Host:     connP.Host,
		Port:     connP.Port,
		Istls:    connP.Istls,
	})
	if err != nil {
		return ds, err
	}
	termc, err := db.FindOrCreateCollection(collP.Term, &driver.CreateCollectionOptions{})
	if err != nil {
		return ds, err
	}
	relc, err := db.FindOrCreateCollection(
		collP.Relationship,
		&driver.CreateCollectionOptions{Type: driver.CollectionTypeEdge},
	)
	if err != nil {
		return ds, err
	}
	graphc, err := db.FindOrCreateCollection(collP.GraphInfo, &driver.CreateCollectionOptions{})
	if err != nil {
		return ds, err
	}
	obog, err := db.FindOrCreateGraph(
		collP.OboGraph,
		[]driver.EdgeDefinition{{
			Collection: relc.Name(),
			From:       []string{termc.Name()},
			To:         []string{termc.Name()},
		}},
	)
	if err != nil {
		return ds, err
	}
	return &arangoSource{
		sess:     sess,
		database: db,
		termc:    termc,
		relc:     relc,
		graphc:   graphc,
		obog:     obog,
	}, nil
}

type arangoSource struct {
	sess     *manager.Session
	database *manager.Database
	termc    driver.Collection
	relc     driver.Collection
	graphc   driver.Collection
	obog     driver.Graph
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
	count, err := a.database.Count(
		fmt.Sprintf(getd, a.graphc.Name(), g.ID()),
	)
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

func (a *arangoSource) UpdateOboGraphInfo(g graph.OboGraph) error {
	key, err := a.graphDocKey(g)
	if err != nil {
		return err
	}
	dg := dbGraphInfo{
		Metadata: a.todbGraphMeta(g),
	}
	_, err = a.graphc.UpdateDocument(
		driver.WithSilent(context.Background()),
		key,
		dg,
	)
	if err != nil {
		return fmt.Errorf("error in updating the graph %s", err)
	}
	return nil
}

// SaveorUpdateTerms insert and update terms in the storage
// and returns no of new and updated terms
func (a *arangoSource) SaveOrUpdateTerms(g graph.OboGraph) (int, int, error) {
	ucount := 0
	icount := 0
	id, err := a.graphDocId(g)
	if err != nil {
		return icount, ucount, err
	}
	tmpColl, err := a.database.CreateCollection(
		generate.RandString(13),
		&driver.CreateCollectionOptions{},
	)
	if err != nil {
		return icount, ucount, err
	}
	defer func() {
		if err := tmpColl.Remove(context.Background()); err != nil {
			log.Printf("error in removing tmp collection %s", err)
		}
	}()
	var dbterms []*dbTerm
	for _, t := range g.Terms() {
		dbterms = append(dbterms, a.todbTerm(id, t))
	}
	_, err = tmpColl.ImportDocuments(
		context.Background(),
		dbterms,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return icount, ucount, err
	}
	// update terms
	ru, err := a.database.Run(termUpdate(
		g.ID(),
		a.graphc.Name(),
		a.termc.Name(),
		tmpColl.Name(),
	))
	if err != nil {
		return icount, ucount, fmt.Errorf("unable to run term update query %s", err)
	}
	if err := ru.Read(&ucount); err != nil {
		return icount, ucount, fmt.Errorf("error in reading number of updates %s", err)
	}
	//insert new terms
	ri, err := a.database.DoRun(tinst, map[string]interface{}{
		"graph_id":          g.ID(),
		"term_collection":   a.termc.Name(),
		"@graph_collection": a.graphc.Name(),
		"@term_collection":  a.termc.Name(),
		"@temp_collection":  tmpColl.Name(),
	})
	if err != nil {
		return icount, ucount, fmt.Errorf("unable to run term insert query %s", err)
	}
	if err := ri.Read(&icount); err != nil {
		return icount, ucount, fmt.Errorf("error in reading number of inserts %s", err)
	}
	return icount, ucount, nil
}

// SaveNewRelationships saves only the new relationships that are absent in the storage
func (a *arangoSource) SaveNewRelationships(g graph.OboGraph) (int, error) {
	ncount := 0
	tmpColl, err := a.database.CreateCollection(
		generate.RandString(12),
		&driver.CreateCollectionOptions{
			Type: driver.CollectionTypeEdge,
		},
	)
	if err != nil {
		return ncount, err
	}
	defer func() {
		if err := tmpColl.Remove(context.Background()); err != nil {
			log.Printf("error in removing tmp collection %s", err)
		}
	}()
	var dbrs []*dbRelationship
	for _, r := range g.Relationships() {
		dbrel, err := a.todbRelationhip(r)
		if err != nil {
			return 0, err
		}
		dbrs = append(dbrs, dbrel)
	}
	_, err = tmpColl.ImportDocuments(
		context.Background(),
		dbrs,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return ncount,
			fmt.Errorf(
				"error in inserting relationships in temp collection %s %s",
				tmpColl.Name(), err,
			)
	}
	r, err := a.database.Run(relInsert(
		g.ID(),
		a.graphc.Name(),
		a.termc.Name(),
		a.relc.Name(),
		a.obog.Name(),
		tmpColl.Name(),
	))
	if err != nil {
		return ncount, fmt.Errorf("unable to run new relationships insert query %s", err)
	}
	if err := r.Read(&ncount); err != nil {
		return ncount, fmt.Errorf("error in reading in no of relationship insert %s", err)
	}
	return ncount, nil
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
	return a.graphDocQuery(
		fmt.Sprintf(getq, a.termc.Name(), string(nid), "d._id"),
	)
}

func (a *arangoSource) graphDocId(g graph.OboGraph) (string, error) {
	return a.graphDocQuery(
		fmt.Sprintf(getq, a.graphc.Name(), g.ID(), "d._id"),
	)
}

func (a *arangoSource) graphDocKey(g graph.OboGraph) (string, error) {
	return a.graphDocQuery(
		fmt.Sprintf(getq, a.graphc.Name(), g.ID(), "d._key"),
	)
}

func (a *arangoSource) graphDocQuery(query string) (string, error) {
	var ret string
	res, err := a.database.Get(query)
	if err != nil {
		return ret, err
	}
	if res.IsEmpty() {
		return ret, errors.New("graph id is absent from database")
	}
	err = res.Read(&ret)
	return ret, err
}

func (a *arangoSource) termUpdate(gname, gcoll, tcoll, temp string) string {
	return fmt.Sprintf(
		tupdt, gcoll, gname, tcoll,
		temp, temp, tcoll, tcoll,
	)
}

func (a *arangoSource) relInsert(gname, gcoll, tcoll, rcoll, graph, temp string) string {
	return fmt.Sprintf(
		rinst, gcoll, tcoll, gname, temp,
		graph, temp, tcoll, rcoll,
	)
}

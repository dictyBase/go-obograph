package arangodb

import (
	"context"
	"errors"
	"fmt"
	"log"

	driver "github.com/arangodb/go-driver"
	manager "github.com/dictyBase/arangomanager"
	"github.com/dictyBase/go-obograph/generate"
	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
	"github.com/go-playground/validator/v10"
)

const seedNo = 13

type removeTempCollection func(tmp *arangoCollection)

// ConnectParams are the parameters required for connecting to arangodb.
type ConnectParams struct {
	User     string `validate:"required"`
	Pass     string `validate:"required"`
	Database string `validate:"required"`
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	Istls    bool
}

// CollectionParams are the arangodb collections required for storing
// OBO graphs.
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

func validateAll(ai ...interface{}) error {
	validate := validator.New()
	for _, iface := range ai {
		if err := validate.Struct(iface); err != nil {
			return fmt.Errorf("error in validating struct %s", err)
		}
	}

	return nil
}

func NewDataSourceFromDb(dbh *manager.Database, collP *CollectionParams) (storage.DataSource, error) {
	if err := validateAll(collP); err != nil {
		return &arangoSource{}, err
	}
	col, err := CreateCollection(dbh, collP)
	if err != nil {
		return &arangoSource{}, err
	}

	return &arangoSource{
		database: dbh,
		termc:    col.Term,
		relc:     col.Rel,
		graphc:   col.Cv,
		obog:     col.Obog,
	}, nil
}

func NewDataSource(connP *ConnectParams, collP *CollectionParams) (storage.DataSource, error) {
	var dsa *arangoSource
	if err := validateAll(connP, collP); err != nil {
		return dsa, err
	}
	_, dbh, err := manager.NewSessionDb(&manager.ConnectParams{
		User:     connP.User,
		Pass:     connP.Pass,
		Database: connP.Database,
		Host:     connP.Host,
		Port:     connP.Port,
		Istls:    connP.Istls,
	})
	if err != nil {
		return dsa, err
	}
	col, err := CreateCollection(dbh, collP)
	if err != nil {
		return dsa, err
	}

	return &arangoSource{
		database: dbh,
		termc:    col.Term,
		relc:     col.Rel,
		graphc:   col.Cv,
		obog:     col.Obog,
	}, nil
}

type arangoSource struct {
	database *manager.Database
	termc    driver.Collection
	relc     driver.Collection
	graphc   driver.Collection
	obog     driver.Graph
}

// SaveOboGraphInfo perist OBO graphs metadata in the storage.
func (a *arangoSource) SaveOboGraphInfo(g graph.OboGraph) error {
	dbg := dbGraphInfo{
		ID:       g.ID(),
		IRI:      g.IRI(),
		Label:    g.Label(),
		Metadata: a.todbGraphMeta(g),
	}
	ctx := driver.WithSilent(context.Background())
	_, err := a.graphc.CreateDocument(ctx, dbg)

	if err != nil {
		return fmt.Errorf("error in creating document %s", err)
	}

	return nil
}

// ExistOboGraph checks for existence of a particular OBO graph.
func (a *arangoSource) ExistsOboGraph(g graph.OboGraph) bool {
	count, err := a.database.CountWithParams(getd,
		map[string]interface{}{
			"@graph_collection": a.graphc.Name(),
			"graph_id":          g.ID(),
		})
	if err != nil {
		return false
	}
	if count > 0 {
		return true
	}

	return false
}

func (a *arangoSource) SaveTerms(grph graph.OboGraph) (int, error) {
	idg, err := a.graphDocID(grph)
	if err != nil {
		return 0, err
	}
	dbterms := make([]*dbTerm, 0)
	for _, t := range grph.Terms() {
		dbterms = append(dbterms, a.todbTerm(idg, t))
	}
	stat, err := a.termc.ImportDocuments(
		context.Background(),
		dbterms,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return 0, fmt.Errorf("error in importing documents %s", err)
	}

	return int(stat.Created), nil
}

func (a *arangoSource) SaveRelationships(g graph.OboGraph) (int, error) {
	dbrs := make([]*dbRelationship, 0)
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
		return 0, fmt.Errorf("error in importing documents %s", err)
	}

	return int(stat.Created), nil
}

func (a *arangoSource) UpdateTerms(g graph.OboGraph) (int, error) {
	return 0, nil
}

func (a *arangoSource) UpdateOboGraphInfo(grph graph.OboGraph) error {
	key, err := a.graphDocKey(grph)
	if err != nil {
		return err
	}
	dbg := dbGraphInfo{
		Metadata: a.todbGraphMeta(grph),
	}
	_, err = a.graphc.UpdateDocument(
		driver.WithSilent(context.Background()),
		key,
		dbg,
	)
	if err != nil {
		return fmt.Errorf("error in updating the graph %s", err)
	}

	return nil
}

// SaveorUpdateTerms insert and update terms in the storage
// and returns no of new and updated terms.
func (a *arangoSource) SaveOrUpdateTerms(grph graph.OboGraph) (*storage.Stats, error) {
	stats := new(storage.Stats)
	id, err := a.graphDocID(grph)
	if err != nil {
		return stats, err
	}
	tmpColl, fn, err := a.loadTermsinTemp(id, grph)
	if err != nil {
		return stats, err
	}
	defer fn(tmpColl)

	return a.manageTerms(grph, tmpColl)
}

func (a *arangoSource) manageTerms(grph graph.OboGraph, tmpColl driver.AccessTarget) (*storage.Stats, error) {
	stats := new(storage.Stats)
	ucount, err := a.editTerms(tupdt, grph, tmpColl)
	if err != nil {
		return stats, err
	}
	icount, err := a.editTerms(tinst, grph, tmpColl)
	if err != nil {
		return stats, err
	}
	ocount, err := a.editTerms(tdelt, grph, tmpColl)
	if err != nil {
		return stats, err
	}
	stats.Created = icount
	stats.Updated = ucount
	stats.Deleted = ocount

	return stats, nil
}

// SaveNewRelationships saves only the new relationships that are absent in the storage.
func (a *arangoSource) SaveNewRelationships(grph graph.OboGraph) (int, error) {
	ncount := 0
	tmpColl, fn, err := a.loadRelationsinTemp(grph)
	if err != nil {
		return ncount, err
	}
	defer fn(tmpColl)
	runner, err := a.database.DoRun(rinst, map[string]interface{}{
		"@relationship_collection": a.relc.Name(),
		"@graph_collection":        a.graphc.Name(),
		"@term_collection":         a.termc.Name(),
		"@temp_collection":         tmpColl.Name(),
		"cvterm_graph":             a.obog.Name(),
		"graph_id":                 grph.ID(),
	})
	if err != nil {
		return ncount, fmt.Errorf("unable to run new relationships insert query %s", err)
	}
	if err := runner.Read(&ncount); err != nil {
		return ncount, fmt.Errorf("error in reading in no of relationship insert %s", err)
	}

	return ncount, nil
}

func (a *arangoSource) todbGraphMeta(grph graph.OboGraph) *dbGraphMeta {
	dpg := make([]*dbGraphProps, 0)
	for _, p := range grph.Meta().BasicPropertyValues() {
		dpg = append(dpg, &dbGraphProps{
			Pred:  p.Pred(),
			Value: p.Value(),
			Curie: curieMap[p.Pred()],
		})
	}

	return &dbGraphMeta{
		Namespace:  grph.Meta().Namespace(),
		Version:    grph.Meta().Version(),
		Properties: dpg,
	}
}

func (a *arangoSource) todbTerm(idn string, trm graph.Term) *dbTerm {
	dbt := &dbTerm{
		ID:         string(trm.ID()),
		Iri:        trm.IRI(),
		Label:      trm.Label(),
		RdfType:    trm.RdfType(),
		Deprecated: trm.IsDeprecated(),
		GraphID:    idn,
	}
	if !trm.HasMeta() {
		return dbt
	}
	dbm := &dbTermMeta{}
	dbm.Properties = termMetaProperties(trm)
	if len(trm.Meta().Xrefs()) > 0 {
		var dbx []*dbMetaXref
		for _, r := range trm.Meta().Xrefs() {
			dbx = append(dbx, &dbMetaXref{Value: r.Value()})
		}
		dbm.Xrefs = dbx
	}
	if len(trm.Meta().Synonyms()) > 0 {
		var dbs []*dbMetaSynonym
		for _, syn := range trm.Meta().Synonyms() {
			dbs = append(dbs, &dbMetaSynonym{
				Value:   syn.Value(),
				Pred:    syn.Pred(),
				Scope:   syn.Scope(),
				IsExact: syn.IsExact(),
				Xrefs:   syn.Xrefs(),
			})
		}
		dbm.Synonyms = dbs
	}
	if len(trm.Meta().Comments()) > 0 {
		dbm.Comments = trm.Meta().Comments()
	}
	if len(trm.Meta().Subsets()) > 0 {
		dbm.Subsets = trm.Meta().Subsets()
	}
	if trm.Meta().Definition() != nil {
		dbm.Definition = &dbMetaDefinition{
			Value: trm.Meta().Definition().Value(),
			Xrefs: trm.Meta().Definition().Xrefs(),
		}
	}
	dbm.Namespace = trm.Meta().Namespace()
	dbt.Metadata = dbm

	return dbt
}

func (a *arangoSource) todbRelationhip(rgp graph.Relationship) (*dbRelationship, error) {
	oMap := make(map[graph.NodeID]string)
	dbr := &dbRelationship{}
	if v, ok := oMap[rgp.Object()]; ok {
		dbr.From = v
	} else {
		id, err := a.getDocID(rgp.Object())
		if err != nil {
			return dbr, err
		}
		oMap[rgp.Object()] = id
		dbr.From = id
	}
	if v, ok := oMap[rgp.Subject()]; ok {
		dbr.To = v
	} else {
		id, err := a.getDocID(rgp.Subject())
		if err != nil {
			return dbr, err
		}
		oMap[rgp.Subject()] = id
		dbr.To = id
	}
	if v, ok := oMap[rgp.Predicate()]; ok {
		dbr.Predicate = v
	} else {
		id, err := a.getDocID(rgp.Predicate())
		if err != nil {
			return dbr, err
		}
		oMap[rgp.Predicate()] = id
		dbr.Predicate = id
	}

	return dbr, nil
}

func (a *arangoSource) getDocID(nid graph.NodeID) (string, error) {
	return a.graphDocQuery(
		getid,
		map[string]interface{}{
			"@db_collection": a.termc.Name(),
			"db_id":          string(nid),
		})
}

func (a *arangoSource) graphDocID(g graph.OboGraph) (string, error) {
	return a.graphDocQuery(
		getid,
		map[string]interface{}{
			"@db_collection": a.graphc.Name(),
			"db_id":          g.ID(),
		})
}

func (a *arangoSource) graphDocKey(g graph.OboGraph) (string, error) {
	return a.graphDocQuery(
		getkey,
		map[string]interface{}{
			"@db_collection": a.graphc.Name(),
			"db_id":          g.ID(),
		})
}

func (a *arangoSource) graphDocQuery(query string, bindVars map[string]interface{}) (string, error) {
	var ret string
	res, err := a.database.GetRow(query, bindVars)
	if err != nil {
		return ret, err
	}
	if res.IsEmpty() {
		return ret, errors.New("graph id is absent from database")
	}
	err = res.Read(&ret)

	return ret, err
}

func (a *arangoSource) editTerms(query string, g graph.OboGraph, tmpColl driver.AccessTarget) (int, error) {
	var ocount int
	runner, err := a.database.DoRun(query, map[string]interface{}{
		"graph_id":          g.ID(),
		"@graph_collection": a.graphc.Name(),
		"@term_collection":  a.termc.Name(),
		"@temp_collection":  tmpColl.Name(),
	})
	if err != nil {
		return ocount, fmt.Errorf("unable to run term query %s", err)
	}
	if err := runner.Read(&ocount); err != nil {
		return ocount, fmt.Errorf("error in reading from database %s", err)
	}

	return ocount, nil
}
func (a *arangoSource) loadRelationsinTemp(grph graph.OboGraph) (*arangoCollection, removeTempCollection, error) {
	coll := new(arangoCollection)
	fnc := func(tmpColl *arangoCollection) {
		if err := tmpColl.Remove(context.Background()); err != nil {
			log.Printf("error in removing tmp collection %s", err)
		}
	}
	rnd, err := generate.RandString(seedNo)
	if err != nil {
		return coll, fnc, fmt.Errorf("error in generating random string %s", err)
	}
	tmpColl, err := a.database.CreateCollection(
		rnd, &driver.CreateCollectionOptions{
			Type: driver.CollectionTypeEdge,
		},
	)
	if err != nil {
		return coll, fnc, err
	}
	dbrs := make([]*dbRelationship, 0)
	for _, r := range grph.Relationships() {
		dbrel, err := a.todbRelationhip(r)
		if err != nil {
			return coll, fnc, err
		}
		dbrs = append(dbrs, dbrel)
	}
	_, err = tmpColl.ImportDocuments(
		context.Background(),
		dbrs,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return coll, fnc, fmt.Errorf("error in importing document %s", err)
	}
	coll.Collection = tmpColl

	return coll, fnc, nil
}

func (a *arangoSource) loadTermsinTemp(idt string, grph graph.OboGraph) (*arangoCollection, removeTempCollection, error) {
	coll := new(arangoCollection)
	fnc := func(tmpColl *arangoCollection) {
		if err := tmpColl.Remove(context.Background()); err != nil {
			log.Printf("error in removing tmp collection %s", err)
		}
	}
	rnd, err := generate.RandString(seedNo)
	if err != nil {
		return coll, fnc, fmt.Errorf("error in generating random string %s", err)
	}
	tmpColl, err := a.database.CreateCollection(
		rnd, &driver.CreateCollectionOptions{},
	)
	if err != nil {
		return coll, fnc, fmt.Errorf("error in creating collection %s", err)
	}
	dbterms := make([]*dbTerm, 0)
	for _, t := range grph.Terms() {
		dbterms = append(dbterms, a.todbTerm(idt, t))
	}
	_, err = tmpColl.ImportDocuments(
		context.Background(),
		dbterms,
		&driver.ImportDocumentOptions{Complete: true},
	)
	if err != nil {
		return coll, fnc, fmt.Errorf("error in importing documents %s", err)
	}
	coll.Collection = tmpColl

	return coll, fnc, nil
}

func termMetaProperties(trm graph.Term) []*dbGraphProps {
	dps := make([]*dbGraphProps, 0)
	if len(trm.Meta().BasicPropertyValues()) == 0 {
		return dps
	}
	for _, prop := range trm.Meta().BasicPropertyValues() {
		dps = append(dps, &dbGraphProps{
			Pred:  prop.Pred(),
			Value: prop.Value(),
			Curie: curieMap[prop.Pred()],
		})
	}

	return dps
}

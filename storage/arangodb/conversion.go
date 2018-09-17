package arangodb

import (
	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage/arangodb/manager"
)

var sMap map[graph.NodeID]string = make(map[graph.NodeID]string)
var pMap map[graph.NodeID]string = make(map[graph.NodeID]string)
var oMap map[graph.NodeID]string = make(map[graph.NodeID]string)

func todbTerm(t graph.Term) *dbTerm {
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

func todbRelationhip(db *manager.Database, r graph.Relationship) (*db.Relationship, error) {
	dbr := &dbRelationship{}
}

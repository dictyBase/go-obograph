package arangodb

import (
	"time"
)

var curieMap = map[string]string{
	"http://www.geneontology.org/formats/oboInOwl#date":                "date",
	"http://www.geneontology.org/formats/oboInOwl#saved-by":            "saved_by",
	"http://www.geneontology.org/formats/oboInOwl#auto-generated-by":   "generated_by",
	"http://www.geneontology.org/formats/oboInOwl#default-namespace":   "namespace",
	"http://www.geneontology.org/formats/oboInOwl#hasOBOFormatVersion": "oboFormat",
	"http://www.geneontology.org/formats/oboInOwl#hasOBONamespace":     "namespace",
	"http://www.geneontology.org/formats/oboInOwl#creation_date":       "date",
	"http://www.w3.org/2000/01/rdf-schema#comment":                     "comment",
	"http://www.geneontology.org/formats/oboInOwl#created_by":          "created_by",
	"http://www.w3.org/2002/07/owl#deprecated":                         "deprecated",
	"http://purl.obolibrary.org/obo/IAO_0100001":                       "replaced_by",
}

type dbGraphInfo struct {
	id        string       `json:"id"`
	iri       string       `json:"iri"`
	label     string       `json:"label"`
	createdAt time.Time    `json:"created_at"`
	updatedAt time.Time    `json:"updated_at"`
	metadata  *dbGraphMeta `json:"metadata"`
}

type dbGraphMeta struct {
	namespace  string          `json:"namespace"`
	version    string          `json:"version"`
	properties []*dbGraphProps `json:"properties"`
}

type dbGraphProps struct {
	pred  string `json:"pred"`
	value string `json:"value"`
	curie string `json:"curie"`
}

type dbTerm struct {
	id       string      `json:"id"`
	iri      string      `json:"iri"`
	label    string      `json:"label"`
	rdfType  string      `json:"rdftype"`
	metadata *dbTermMeta `json:"metadata"`
}

type dbTermMeta struct {
	namespace  string            `json:"namespace"`
	comments   []string          `json:"comments"`
	subsets    []string          `json:"subsets"`
	definition *dbMetaDefinition `json:"definition"`
	synonyms   []*dbMetaSynonym  `json:"synonyms"`
	xrefs      []*dbMetaXref     `json:"xrefs"`
	properties []*dbGraphProps   `json:"properties"`
}

type dbMetaDefinition struct {
	value string   `json:"value"`
	xrefs []string `json:"xrefs"`
}

type dbMetaSynonym struct {
	value   string   `json:"value"`
	pred    string   `json:"pred"`
	scope   string   `json:"scope"`
	isExact bool     `json:"is_exact"`
	xrefs   []string `json:"xrefs"`
}

type dbMetaXref struct {
	value string `json:"value"`
}

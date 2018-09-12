package arangodb

import (
	"time"
)

var curieMap = map[string]string{
	"http://www.geneontology.org/formats/oboInOwl#date":                "date",
	"http://www.geneontology.org/formats/oboInOwl#saved-by":            "savedBy",
	"http://www.geneontology.org/formats/oboInOwl#auto-generated-by":   "generatedBy",
	"http://www.geneontology.org/formats/oboInOwl#default-namespace":   "namespace",
	"http://www.geneontology.org/formats/oboInOwl#hasOBOFormatVersion": "oboFormat",
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
	namespace  `json:"namespace"`
	version    `json:"version"`
	properties []*dbGraphProps `json:"properties"`
}

type dbGraphProps struct {
	pred  string `json:"pred"`
	value string `json:"value"`
	curie string `json:"curie"`
}

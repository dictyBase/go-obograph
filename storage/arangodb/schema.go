package arangodb

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
	"http://purl.org/dc/elements/1.1/description":                      "description",
	"http://purl.org/dc/terms/license":                                 "license",
	"http://purl.org/dc/elements/1.1/title":                            "title",
}

type dbGraphInfo struct {
	Id       string       `json:"id,omitempty"`
	IRI      string       `json:"iri,omitempty"`
	Label    string       `json:"label,omitempty"`
	Metadata *dbGraphMeta `json:"metadata,omitempty"`
}

type dbGraphMeta struct {
	Namespace  string          `json:"namespace"`
	Version    string          `json:"version"`
	Properties []*dbGraphProps `json:"properties,omitempty"`
}

type dbGraphProps struct {
	Pred  string `json:"pred,omitempty"`
	Value string `json:"value,omitempty"`
	Curie string `json:"curie,omitempty"`
}

type dbTerm struct {
	GraphId    string      `json:"graph_key"`
	Id         string      `json:"id"`
	Iri        string      `json:"iri"`
	Label      string      `json:"label"`
	RdfType    string      `json:"rdftype"`
	Deprecated bool        `json:"deprecated"`
	Metadata   *dbTermMeta `json:"metadata"`
}

type dbTermMeta struct {
	Namespace  string            `json:"namespace"`
	Comments   []string          `json:"comments"`
	Subsets    []string          `json:"subsets"`
	Definition *dbMetaDefinition `json:"definition"`
	Synonyms   []*dbMetaSynonym  `json:"synonyms"`
	Xrefs      []*dbMetaXref     `json:"xrefs"`
	Properties []*dbGraphProps   `json:"properties"`
}

type dbMetaDefinition struct {
	Value string   `json:"value"`
	Xrefs []string `json:"xrefs"`
}

type dbMetaSynonym struct {
	Value   string   `json:"value"`
	Pred    string   `json:"pred"`
	Scope   string   `json:"scope"`
	IsExact bool     `json:"is_exact"`
	Xrefs   []string `json:"xrefs"`
}

type dbMetaXref struct {
	Value string `json:"value"`
}

type dbRelationship struct {
	From      string `json:"_from"`
	To        string `json:"_to"`
	Predicate string `json:"predicate"`
}

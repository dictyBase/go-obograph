// Package schema provides type definitions for decoding OBO Graphs in JSON
// format
package schema

// OboJSON models the entire JSON schema of OBO Graph
type OboJSON struct {
	Graphs []*OboJSONGraph `json:"graphs"`
}

// OboJSONGraph models the graph section of OBO graph
type OboJSONGraph struct {
	ID                  string        `json:"id"`
	Edges               []*JSONEdge   `json:"edges"`
	Nodes               []*JSONNode   `json:"nodes"`
	Meta                *JSONMeta     `json:"meta"`
	EquivalentNodesSets []interface{} `json:"equivalentNodesSets"`
	DomainRangeAxioms   []interface{} `json:"domainRangeAxioms"`
	PropertyChainAxioms []struct {
		ChainPredicateIds []string `json:"chainPredicateIds"`
		PredicateID       string   `json:"predicateId"`
	} `json:"propertyChainAxioms"`
	LogicalDefinitionAxioms []struct {
		DefinedClassID string   `json:"definedClassId"`
		GenusIds       []string `json:"genusIds"`
		Restrictions   []struct {
			FillerID   string `json:"fillerId"`
			PropertyID string `json:"propertyId"`
		} `json:"restrictions"`
	} `json:"logicalDefinitionAxioms"`
}

// JSONMeta models the meta section of OBO graph
type JSONMeta struct {
	BasicPropertyValues []*JSONProperty `json:"basicPropertyValues"`
	Subsets             []string        `json:"subsets"`
	Version             string          `json:"version"`
	Synonyms            []*JSONSynonym  `json:"synonyms"`
	Comments            []string        `json:"comments"`
	Definition          struct {
		Val   string   `json:"val"`
		Xrefs []string `json:"xrefs"`
	} `json:"definition"`
	Xrefs []struct {
		Val string `json:"val"`
	} `json:"xrefs"`
}

// JSONEdge models the edges of OBO graph
type JSONEdge struct {
	Obj  string `json:"obj"`
	Pred string `json:"pred"`
	Sub  string `json:"sub"`
}

// JSONNode models the nodes of OBO graph
type JSONNode struct {
	ID       string    `json:"id"`
	Lbl      string    `json:"lbl"`
	Meta     *JSONMeta `json:"meta"`
	JSONType string    `json:"type"`
}

// JSONSynonym models the synonyms of the nodes
type JSONSynonym struct {
	Pred  string   `json:"pred"`
	Val   string   `json:"val"`
	Xrefs []string `json:"xrefs"`
}

// JSONProperty models the properties of the nodes
type JSONProperty struct {
	Pred string `json:"pred"`
	Val  string `json:"val"`
}

//type DbTerm struct {
//Id         string           `json:"id"`
//IRI        string           `json:"iri"`
//Label      string           `json:"label"`
//ChadoId    string           `json:"chado-id"`
//Type       string           `json:"type"`
//Definition string           `json:"definition"`
//Synonyms   []*DbTermSynonym `json:"synonyms"`
//Dbxrefs    []*DbTermDbxref  `json:"dbxrefs"`
//Property   *DbTermProp      `json:"property"`
//}

//type DbTermDbxref struct {
//Database  string `json:"database"`
//Accession string `json:"accession"`
//}

//type DbTermSynonym struct {
//Name  string `json:"name"`
//Scope string `json:"scope"`
//}

//type DbTermProp struct {
//Namespace  string    `json:"namespace"`
//ReplacedBy string    `json:"replaced_by"`
//Consider   []string  `json:"consider"`
//Deprecated bool      `json:"deprecated"`
//Comment    string    `json:"comment"`
//Value      string    `json:"value"`
//CreatedBy  string    `json:"created_by"`
//CreatedOn  time.Time `json:"created_on"`
//}

//type DbRelationship struct {
//From  string `json:"_from,omitempty"`
//To    string `json:"_to,omitempty"`
//Id    string `json:"id,omitempty"`
//Label string `json:"label,omitempty"`
//}

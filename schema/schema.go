// Package schema provides type definitions for decoding OBO Graphs in JSON
// format
package schema

// OboJSON models the entire JSON schema of OBO Graph.
type OboJSON struct {
	Graphs []*OboJSONGraph `json:"graphs"`
}

// OboJSONGraph models the graph section of OBO graph.
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

// JSONMeta models the meta section of OBO graph.
type JSONMeta struct {
	BasicPropertyValues []*JSONProperty `json:"basicPropertyValues"`
	Synonyms            []*JSONSynonym  `json:"synonyms"`
	Subsets             []string        `json:"subsets"`
	Comments            []string        `json:"comments"`
	Definition          *JSONDefintion  `json:"definition"`
	Version             string          `json:"version"`
	Deprecated          bool            `json:"deprecated"`
	Xrefs               []struct {
		Val string `json:"val"`
	} `json:"xrefs"`
}

// JSONDefintion models the definition subsection of meta section.
type JSONDefintion struct {
	Val   string   `json:"val"`
	Xrefs []string `json:"xrefs"`
}

// JSONEdge models the edges of OBO graph.
type JSONEdge struct {
	Obj  string `json:"obj"`
	Pred string `json:"pred"`
	Sub  string `json:"sub"`
}

// JSONNode models the nodes of OBO graph.
type JSONNode struct {
	ID       string    `json:"id"`
	Lbl      string    `json:"lbl"`
	Meta     *JSONMeta `json:"meta"`
	JSONType string    `json:"type"`
}

// JSONSynonym models the synonyms of the nodes.
type JSONSynonym struct {
	Pred  string   `json:"pred"`
	Val   string   `json:"val"`
	Xrefs []string `json:"xrefs"`
}

// JSONProperty models the properties of the nodes.
type JSONProperty struct {
	Pred string `json:"pred"`
	Val  string `json:"val"`
}

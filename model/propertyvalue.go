package model

//PropertyValue for modeling ontology metadata other than
//term and their relationships
type PropertyValue struct {
	val  string
	refs []string
	prd  string
}

// Value is the value of property
func (p *PropertyValue) Value() string {
	return p.val
}

// Xrefs is list of data supporting the property-value assertion
func (p *PropertyValue) Xrefs() []string {
	return p.refs
}

// Pred corresponds to OWL properties
func (p *PropertyValue) Pred() string {
	return p.prd
}

// BasicPropertyValue is a generic PropertyValue
type BasicPropertyValue struct {
	*PropertyValue
}

// NewBasicPropertyValue returns a new Basic
func NewBasicPropertyValue(prd, val string) *BasicPropertyValue {
	return &BasicPropertyValue{
		&PropertyValue{
			val: val,
			prd: prd,
		},
	}
}

// Definition represents a textual definition of an ontology term
type Definition struct {
	*PropertyValue
}

// NewDefinition returns a new Definition
func NewDefinition(val string, refs []string) *Definition {
	return &Definition{&PropertyValue{val: val, refs: refs}}

}

// Synonym represent an alternate term for the node
type Synonym struct {
	*PropertyValue
}

// NewSynonymWithRefs returns a new Synonym
func NewSynonymWithRefs(prd, val string, refs []string) *Synonym {
	return &Synonym{
		&PropertyValue{
			val:  val,
			prd:  prd,
			refs: refs,
		},
	}
}

// NewSynonym returns a new Synonym
func NewSynonym(prd, val string) *Synonym {
	return &Synonym{
		&PropertyValue{
			val: val,
			prd: prd,
		},
	}
}

// IsExact is a convenience method to check for EXACT scope
func (s *Synonym) IsExact() bool {
	if s.Pred() == "hasExactSynonym" {
		return true
	}
	return false
}

// Scope returns OBO-style scope of synonym
func (s *Synonym) Scope() string {
	scope := "RELATED"
	switch s.Pred() {
	case "hasExactSynonym":
		scope = "EXACT"
		break
	case "hasNarrowSynonym":
		scope = "NARROW"
		break
	case "hasBroadSynonym":
		scope = "BROAD"
		break
	default:
		break
	}
	return scope
}

// Xref support the property-value assertion
type Xref struct {
	*PropertyValue
}

// NewXref returns a new Xref
func NewXref(val string) *Xref {
	return &Xref{
		&PropertyValue{
			val: val,
		},
	}
}

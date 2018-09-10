package graph

import (
	"github.com/dictyBase/go-obograph/model"
)

// Relationship is an interface for representing relationship
// between terms
type Relationship interface {
	// Object is the unique identifier for parent term
	Object() NodeID
	// Subject is the unique identifier for child term
	Subject() NodeID
	// Predicate is the unique identifier for term that describes the relationship
	Predicate() NodeID
	// Meta returns the relationship's Meta object
	Meta() *model.Meta
}

type edge struct {
	obj  NodeID
	subj NodeID
	pred NodeID
	meta *model.Meta
}

// NewRelationshipWithMeta is a constructor for Relationship
// that receives an additional Meta object
func NewRelationshipWithMeta(obj, subj, pred NodeID, m *model.Meta) Relationship {
	return &edge{
		obj:  obj,
		subj: subj,
		pred: pred,
		meta: m,
	}
}

// NewRelationship is a constructor for Relationship
func NewRelationship(obj, subj, pred NodeID) Relationship {
	return &edge{
		obj:  obj,
		subj: subj,
		pred: pred,
	}
}

// Meta returns the relationship's Meta object
func (e *edge) Meta() *model.Meta {
	return e.meta
}

// Subject is the unique identifier for child term
func (e *edge) Subject() NodeID {
	return e.subj
}

// Predicate is the unique identifier for term that describes the relationship
func (e *edge) Predicate() NodeID {
	return e.pred
}

// Object is the unique identifier for parent term
func (e *edge) Object() NodeID {
	return e.obj
}

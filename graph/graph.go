package graph

import (
	"fmt"

	"github.com/dictyBase/go-obograph/model"
)

// NodeID is a custom type for holding a node id.
type NodeID string

// OboGraph is an interface for accessing OBO Graphs.
type OboGraph interface {
	// IRI represents a stable URL for locating the source OWL formatted file
	IRI() string
	// ID is a short and unique name of the graph
	ID() string
	// Label is a short human readable description of the graph
	Label() string
	// Meta returns the associated Meta container
	Meta() *model.Meta
	// ExistsTerm checks for existence of a term
	ExistsTerm(NodeID) bool
	// GetTerm fetches an existing term
	GetTerm(NodeID) Term
	// GetRelationship fetches relationship(edge) between parent(object) and
	// children(subject)
	GetRelationship(NodeID, NodeID) Relationship
	// Relationships returns all relationships(edges) in the graph
	Relationships() []Relationship
	// Terms returns all terms(node/vertex) in the graph
	Terms() []Term
	// TermsByType provides a filtered list of specific terms
	TermsByType(string) []Term
	// Children returns all children terms(depth one)
	Children(NodeID) []Term
	// Parents returns all parent terms(depth one)
	Parents(NodeID) []Term
	// Ancestors returns all reachable(direct or indirect) parent terms. It uses
	// BFS algorithm
	Ancestors(NodeID) []Term
	// Descendents returns all reachable(direct or indirect) children terms. It uses
	// BFS algorithm
	Descendents(NodeID) []Term
	// DescendentsDFS returns all reachable(direct or indirect) children terms
	// using DFS algorithm.
	DescendentsDFS(NodeID) []Term
	// AddRelationship creates relationship between terms, it overrides the
	// existing terms and relationship
	AddRelationship(Term, Term, Term) error
	// AddRelationshipWithID creates relationship between existing terms
	AddRelationshipWithID(NodeID, NodeID, NodeID) error
	// AddTerm add a new Term to the graph overwriting any existing one
	AddTerm(Term)
}

type graph struct {
	nodes     map[NodeID]Term
	edgesDown map[NodeID]map[NodeID]Relationship
	edgesUp   map[NodeID]map[NodeID]Relationship
	meta      *model.Meta
	id        string
	lbl       string
	iri       string
}

func newOboGraph(m *model.Meta, idn, iri string) OboGraph {
	return &graph{
		nodes:     make(map[NodeID]Term),
		edgesUp:   make(map[NodeID]map[NodeID]Relationship),
		edgesDown: make(map[NodeID]map[NodeID]Relationship),
		meta:      m,
		id:        idn,
		iri:       iri,
	}
}

// Label is a short human readable description of the graph.
func (g *graph) Label() string {
	return g.lbl
}

// ID is a short and unique name of the graph.
func (g *graph) ID() string {
	return g.id
}

// IRI represents a stable URL for locating the source OWL formatted file.
func (g *graph) IRI() string {
	return g.iri
}

// Meta returns the associated Meta container.
func (g *graph) Meta() *model.Meta {
	return g.meta
}

// Terms returns all terms(node/vertex) in the graph.
func (g *graph) Terms() []Term {
	trm := make([]Term, 0)
	for _, n := range g.nodes {
		trm = append(trm, n)
	}

	return trm
}

// TermsByType provides a filtered list of specific terms.
func (g *graph) TermsByType(rtype string) []Term {
	trm := make([]Term, 0)
	for _, n := range g.nodes {
		if n.RdfType() == rtype {
			trm = append(trm, n)
		}
	}

	return trm
}

// Relationships returns all relationships(edges) in the graph.
func (g *graph) Relationships() []Relationship {
	var rel []Relationship
	for id := range g.edgesDown {
		for k := range g.edgesDown[id] {
			rel = append(rel, g.edgesDown[id][k])
		}
	}

	return rel
}

// Children returns all children terms(depth one).
func (g *graph) Children(id NodeID) []Term {
	return g.getTerms(id, g.edgesDown)
}

// Parents returns all parent terms(depth one).
func (g *graph) Parents(id NodeID) []Term {
	return g.getTerms(id, g.edgesUp)
}

// DescendentsDFS returns all reachable(direct or indirect) children terms
// using DFS algorithm.
func (g *graph) DescendentsDFS(idn NodeID) []Term {
	// slice of descendents
	drm := make([]Term, 0)
	// make sure the node exists in the graph
	if _, ok := g.nodes[idn]; !ok {
		return drm
	}
	// stack of term ids
	stn := make([]NodeID, 0)
	// keep track of visited terms
	visited := make(map[NodeID]bool)

	// push the first term(id)
	stn = append(stn, idn)
	for len(stn) > 0 {
		// get the last term(id)
		nid := stn[len(stn)-1]
		if len(stn) == 1 { // the first case
			stn = stn[:0]
		} else { // remove the last item from stack
			stn = stn[:len(stn)-1]
		}
		// mark them if not visited
		if _, ok := visited[nid]; !ok {
			visited[nid] = true
		}
		// get children of this term
		for _, child := range g.Children(nid) {
			// if not visited push them in the stack
			if _, ok := visited[child.ID()]; !ok {
				drm = append(drm, child)
				stn = append(stn, child.ID())
			}
		}
	}

	return drm
}

// Descendents returns all reachable(direct or indirect) children terms. It uses
// BFS algorithm.
func (g *graph) Descendents(idn NodeID) []Term {
	// slice of descendents
	drm := make([]Term, 0)
	// make sure the node exists in the graph
	if _, ok := g.nodes[idn]; !ok {
		return drm
	}
	// queue of terms
	qid := make([]NodeID, 0)
	// keep track of visited terms
	visited := make(map[NodeID]bool)

	// queue the first item
	qid = append(qid, idn)
	// mark it visited
	visited[idn] = true
	for len(qid) > 0 {
		// dequeue the first element
		nid := qid[0]
		if len(qid) == 1 { // the first case
			qid = qid[:0]
		} else { // remove the first element
			qid = qid[1:]
		}
		// get children of this term
		for _, child := range g.Children(nid) {
			// queue if not visited
			if _, ok := visited[child.ID()]; !ok {
				qid = append(qid, child.ID())
				// mark them visited
				visited[child.ID()] = true
				// collect the children
				drm = append(drm, child)
			}
		}
	}

	return drm
}

// Ancestors returns all reachable(direct or indirect) parent terms. It uses
// BFS algorithm.
func (g *graph) Ancestors(idn NodeID) []Term {
	// slice of ancestors
	var atrm []Term
	// make sure the node exists in the graph
	if _, ok := g.nodes[idn]; !ok {
		return atrm
	}
	// queue of terms
	qid := make([]NodeID, 0)
	// keep track of visited terms
	visited := make(map[NodeID]bool)

	// queue the first item
	qid = append(qid, idn)
	// mark it visited
	visited[idn] = true
	for len(qid) > 0 {
		// dequeue
		nid := qid[len(qid)-1]
		qid = qid[:len(qid)-1]
		// get children of this term
		for _, parent := range g.Parents(nid) {
			// queue if not visited
			if _, ok := visited[parent.ID()]; !ok {
				qid = append(qid, parent.ID())
				// mark them visited
				visited[parent.ID()] = true
				// collect the children
				atrm = append(atrm, parent)
			}
		}
	}

	return atrm
}

// ExistsTerm checks for existence of a term.
func (g *graph) ExistsTerm(id NodeID) bool {
	_, ok := g.nodes[id]

	return ok
}

// GetTerm fetches an existing term.
func (g *graph) GetTerm(id NodeID) Term {
	return g.nodes[id]
}

// GetRelationship fetches relationship(edge) between parent(object) and
// children(subject).
func (g *graph) GetRelationship(obj NodeID, subj NodeID) (rel Relationship) {
	if v, ok := g.edgesDown[obj]; ok {
		if r, ok := v[subj]; ok {
			return r
		}
	}

	return rel
}

// AddTerm add a new Term to the graph overwriting any existing one.
func (g *graph) AddTerm(t Term) {
	g.nodes[t.ID()] = t
}

// AddRelationship creates relationship between terms, it overrides the
// existing terms and relationship.
func (g *graph) AddRelationship(obj, subj, pred Term) error {
	g.nodes[obj.ID()] = obj
	g.nodes[subj.ID()] = subj
	g.nodes[pred.ID()] = pred
	rel := NewRelationship(
		obj.ID(),
		subj.ID(),
		pred.ID(),
	)
	if v, ok := g.edgesDown[obj.ID()]; ok {
		v[subj.ID()] = rel
		g.edgesDown[obj.ID()] = v
	} else {
		g.edgesDown[obj.ID()] = map[NodeID]Relationship{subj.ID(): rel}
	}
	if v, ok := g.edgesUp[subj.ID()]; ok {
		v[obj.ID()] = rel
		g.edgesUp[subj.ID()] = v
	} else {
		g.edgesUp[subj.ID()] = map[NodeID]Relationship{obj.ID(): rel}
	}

	return nil
}

// AddRelationshipWithID creates relationship between existing terms.
func (g *graph) AddRelationshipWithID(obj, subj, pred NodeID) error {
	if _, ok := g.nodes[obj]; !ok {
		return fmt.Errorf("object node id %s does not exist", obj)
	}
	if _, ok := g.nodes[subj]; !ok {
		return fmt.Errorf("subject node id %s does not exist", subj)
	}
	if _, ok := g.nodes[pred]; !ok {
		return fmt.Errorf("predicate node id %s does not exist", pred)
	}
	rel := NewRelationship(
		obj,
		subj,
		pred,
	)
	if v, ok := g.edgesDown[obj]; ok {
		v[subj] = rel
		g.edgesDown[obj] = v
	} else {
		g.edgesDown[obj] = map[NodeID]Relationship{subj: rel}
	}
	if v, ok := g.edgesUp[subj]; ok {
		v[obj] = rel
		g.edgesUp[subj] = v
	} else {
		g.edgesUp[subj] = map[NodeID]Relationship{obj: rel}
	}

	return nil
}

func (g *graph) getTerms(id NodeID, edges map[NodeID]map[NodeID]Relationship) []Term {
	trm := make([]Term, 0)
	if _, ok := g.nodes[id]; ok {
		for nid := range edges[id] {
			trm = append(trm, g.nodes[nid])
		}
	}

	return trm
}

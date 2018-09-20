package graph

import (
	"fmt"
	"strings"
	"time"

	"github.com/dictyBase/go-obograph/model"
)

// NodeID is a custom type for holding a node id
type NodeID string

// OboGraph is an interface for accessing OBO Graphs
type OboGraph interface {
	// IRI represents a stable URL for locating the source OWL formatted file
	IRI() string
	// ID is a short and unique name of the graph
	ID() string
	// Label is a short human readable description of the graph
	Label() string
	// Timestamp provides the date associated with the graph
	Timestamp() time.Time
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

func newOboGraph(m *model.Meta, id, iri string) OboGraph {
	return &graph{
		nodes:     make(map[NodeID]Term),
		edgesUp:   make(map[NodeID]map[NodeID]Relationship),
		edgesDown: make(map[NodeID]map[NodeID]Relationship),
		meta:      m,
		id:        id,
		iri:       iri,
	}
}

// Label is a short human readable description of the graph
func (g *graph) Label() string {
	return g.lbl
}

// ID is a short and unique name of the graph
func (g *graph) ID() string {
	return g.id
}

// IRI represents a stable URL for locating the source OWL formatted file
func (g *graph) IRI() string {
	return g.iri
}

// Meta returns the associated Meta container
func (g *graph) Meta() *model.Meta {
	return g.meta
}

// Terms returns all terms(node/vertex) in the graph
func (g *graph) Terms() []Term {
	var t []Term
	for _, n := range g.nodes {
		t = append(t, n)
	}
	return t
}

// TermsByType provides a filtered list of specific terms
func (g *graph) TermsByType(rtype string) []Term {
	var t []Term
	for _, n := range g.nodes {
		if n.RdfType() == rtype {
			t = append(t, n)
		}
	}
	return t
}

// Relationships returns all relationships(edges) in the graph
func (g *graph) Relationships() []Relationship {
	var rel []Relationship
	for id := range g.edgesDown {
		for k := range g.edgesDown[id] {
			rel = append(rel, g.edgesDown[id][k])
		}
	}
	return rel
}

// Children returns all children terms(depth one)
func (g *graph) Children(id NodeID) []Term {
	var t []Term
	if _, ok := g.nodes[id]; ok {
		for nid := range g.edgesDown[id] {
			t = append(t, g.nodes[nid])
		}
	}
	return t
}

// Parents returns all parent terms(depth one)
func (g *graph) Parents(id NodeID) []Term {
	var t []Term
	if _, ok := g.nodes[id]; ok {
		for nid := range g.edgesUp[id] {
			t = append(t, g.nodes[nid])
		}
	}
	return t
}

// DescendentsDFS returns all reachable(direct or indirect) children terms
// using DFS algorithm.
func (g *graph) DescendentsDFS(id NodeID) []Term {
	// slice of descendents
	var d []Term
	//make sure the node exists in the graph
	if _, ok := g.nodes[id]; !ok {
		return d
	}
	// stack of term ids
	var st []NodeID
	// keep track of visited terms
	visited := make(map[NodeID]bool)

	//push the first term(id)
	st = append(st, id)
	for len(st) > 0 {
		//get the last term(id)
		nid := st[len(st)-1]
		if len(st) == 1 { // the first case
			st = st[:0]
		} else { // remove the last item from stack
			st = st[:len(st)-1]
		}
		// mark them if not visited
		if _, ok := visited[nid]; !ok {
			visited[nid] = true
		}
		// get children of this term
		for _, child := range g.Children(nid) {
			// if not visited push them in the stack
			if _, ok := visited[child.ID()]; !ok {
				d = append(d, child)
				st = append(st, child.ID())
			}
		}
	}
	return d
}

// Descendents returns all reachable(direct or indirect) children terms. It uses
// BFS algorithm
func (g *graph) Descendents(id NodeID) []Term {
	// slice of descendents
	var d []Term
	//make sure the node exists in the graph
	if _, ok := g.nodes[id]; !ok {
		return d
	}
	// queue of terms
	var q []NodeID
	// keep track of visited terms
	visited := make(map[NodeID]bool)

	//queue the first item
	q = append(q, id)
	//mark it visited
	visited[id] = true
	for len(q) > 0 {
		//dequeue the first element
		nid := q[0]
		if len(q) == 1 { // the first case
			q = q[:0]
		} else { // remove the first element
			q = q[1:]
		}
		// get children of this term
		for _, child := range g.Children(nid) {
			// queue if not visited
			if _, ok := visited[child.ID()]; !ok {
				q = append(q, child.ID())
				// mark them visited
				visited[child.ID()] = true
				// collect the children
				d = append(d, child)
			}
		}
	}
	return d
}

// Ancestors returns all reachable(direct or indirect) parent terms. It uses
// BFS algorithm
func (g *graph) Ancestors(id NodeID) []Term {
	// slice of ancestors
	var a []Term
	//make sure the node exists in the graph
	if _, ok := g.nodes[id]; !ok {
		return a
	}
	// queue of terms
	var q []NodeID
	// keep track of visited terms
	visited := make(map[NodeID]bool)

	//queue the first item
	q = append(q, id)
	//mark it visited
	visited[id] = true
	for len(q) > 0 {
		//dequeue
		nid := q[len(q)-1]
		q = q[:len(q)-1]
		// get children of this term
		for _, parent := range g.Parents(nid) {
			// queue if not visited
			if _, ok := visited[parent.ID()]; !ok {
				q = append(q, parent.ID())
				// mark them visited
				visited[parent.ID()] = true
				// collect the children
				a = append(a, parent)
			}
		}
	}
	return a
}

// ExistsTerm checks for existence of a term
func (g *graph) ExistsTerm(id NodeID) bool {
	_, ok := g.nodes[id]
	return ok
}

// GetTerm fetches an existing term
func (g *graph) GetTerm(id NodeID) Term {
	return g.nodes[id]
}

// GetRelationship fetches relationship(edge) between parent(object) and
// children(subject)
func (g *graph) GetRelationship(obj NodeID, subj NodeID) (rel Relationship) {
	if v, ok := g.edgesDown[obj]; ok {
		if r, ok := v[subj]; ok {
			return r
		}
	}
	return rel
}

// AddTerm add a new Term to the graph overwriting any existing one
func (g *graph) AddTerm(t Term) {
	g.nodes[t.ID()] = t
}

// AddRelationship creates relationship between terms, it overrides the
// existing terms and relationship
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

// AddRelationshipWithID creates relationship between existing terms
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

func (g *graph) Timestamp() time.Time {
	//06:21:2018 13:11
	layout := "01:02:2006 15:04"
	for _, p := range g.Meta().BasicPropertyValues() {
		if strings.HasSuffix(p.Pred(), "#date") {
			t, _ := time.Parse(layout, p.Value())
			return t
		}
	}
	return time.Now()
}

func termFilter(terms []Term, fn func(Term) bool) []Term {
	var termf []Term
	for _, t := range terms {
		if fn(t) {
			termf = append(termf, t)
		}
	}
	return termf
}

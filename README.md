# go-obograph

[![License](https://img.shields.io/badge/License-BSD%202--Clause-blue.svg)](LICENSE)  
![GitHub action](https://github.com/dictyBase/go-obograph/workflows/Continuous%20integration/badge.svg)
[![codecov](https://codecov.io/gh/dictyBase/go-obograph/branch/develop/graph/badge.svg)](https://codecov.io/gh/dictyBase/go-obograph)   
[![Funding](https://badgen.net/badge/NIGMS/Rex%20L%20Chisholm,dictyBase,DCR/yellow?list=|)](https://reporter.nih.gov/project-details/10024726)

Go library for reading, traversing, and persisting [OBO Graphs](https://github.com/geneontology/obographs) — the JSON serialization of OWL ontologies used by the Gene Ontology and other biomedical ontologies. Backed by ArangoDB for persistent storage.

## Contents

- [Install](#install)
- [Graph API](#graph-api)
  - [Build a Graph](#build-a-graph)
  - [OboGraph Interface](#obograph-interface)
  - [Traversal](#traversal)
  - [Term](#term)
  - [Relationship](#relationship)
  - [Meta](#meta)
- [Storage](#storage)
  - [DataSource Interface](#datasource-interface)
  - [ArangoDB Backend](#arangodb-backend)
  - [Load from JSON](#load-from-json)
- [CLI Integration](#cli-integration)
  - [Flags](#flags)
  - [CLI Action](#cli-action)

## Install

```bash
go get github.com/dictyBase/go-obograph
```

## Graph API

### Build a Graph

`BuildGraph` parses a JSON-encoded OBO Graph from any `io.Reader` and returns an in-memory `OboGraph` ready for traversal.

```go
import "github.com/dictyBase/go-obograph/graph"

file, _ := os.Open("go.obo.json")
defer file.Close()

grph, err := graph.BuildGraph(file)
```

### OboGraph Interface

The `OboGraph` interface provides read access to nodes (terms) and edges (relationships), plus BFS/DFS traversal and mutation methods.

```go
type OboGraph interface {
    IRI() string
    ID() string
    Label() string
    Meta() *model.Meta
    ExistsTerm(id NodeID) bool
    GetTerm(id NodeID) Term
    GetRelationship(obj NodeID, subj NodeID) Relationship
    Relationships() []Relationship
    Terms() []Term
    TermsByType(rtype string) []Term
    Children(id NodeID) []Term
    Parents(id NodeID) []Term
    Ancestors(id NodeID) []Term
    Descendents(id NodeID) []Term
    DescendentsDFS(id NodeID) []Term
    AddRelationship(obj Term, subj Term, pred Term) error
    AddRelationshipWithID(obj NodeID, subj NodeID, pred NodeID) error
    AddTerm(t Term)
}
```

### Traversal

Direct neighbours and transitive closures:

```go
children := grph.Children(graph.NodeID("GO:0008150"))   // depth-1 children
parents  := grph.Parents(graph.NodeID("GO:0008150"))    // depth-1 parents
ancestors := grph.Ancestors(graph.NodeID("GO:0008150"))  // all reachable parents (BFS)
descBFS   := grph.Descendents(graph.NodeID("GO:0008150")) // all reachable children (BFS)
descDFS   := grph.DescendentsDFS(graph.NodeID("GO:0008150")) // all reachable children (DFS)
```

Filter terms by RDF type:

```go
// CLASS, INDIVIDUAL, or PROPERTY
properties := grph.TermsByType("PROPERTY")
```

### Term

Each node in the graph is a `Term` identified by a `NodeID`.

```go
type Term interface {
    ID() NodeID
    HasMeta() bool
    Meta() *model.Meta
    RdfType() string    // CLASS, INDIVIDUAL, or PROPERTY
    Label() string
    IRI() string
    IsDeprecated() bool
}
```

Construct terms programmatically:

```go
t := graph.NewTerm(
    graph.NodeID("GO:0008150"),
    "CLASS",
    "biological_process",
    "http://purl.obolibrary.org/obo/GO_0008150",
)

tWithMeta := graph.NewTermWithMeta(
    graph.NodeID("GO:0008150"),
    model.NewMeta(&model.MetaOptions{Deprecated: true}),
    "CLASS",
    "biological_process",
    "http://purl.obolibrary.org/obo/GO_0008150",
)
```

### Relationship

Edges connect an object (parent) to a subject (child) via a predicate.

```go
type Relationship interface {
    Object() NodeID
    Subject() NodeID
    Predicate() NodeID
    Meta() *model.Meta
}
```

```go
rel := graph.NewRelationship(
    graph.NodeID("GO:0008150"),   // object (parent)
    graph.NodeID("GO:0009987"),   // subject (child)
    graph.NodeID("is_a"),         // predicate
)

relWithMeta := graph.NewRelationshipWithMeta(
    graph.NodeID("GO:0008150"),
    graph.NodeID("GO:0009987"),
    graph.NodeID("is_a"),
    model.NewMeta(&model.MetaOptions{}),
)
```

### Meta

`Meta` wraps ontology metadata — definitions, synonyms, cross-references, comments, subsets, and deprecation status. Returned by `Term.Meta()` and `Relationship.Meta()`.

```go
meta := term.Meta()
meta.Definition()           // *Definition
meta.Synonyms()            // []*Synonym
meta.Xrefs()               // []*Xref
meta.XrefsValues()         // []string (values only)
meta.Comments()            // []string
meta.Subsets()             // []string
meta.Version()             // string (graph-level only)
meta.Namespace()           // string
meta.IsDeprecated()        // bool
meta.BasicPropertyValues() // []*BasicPropertyValue
```

Construct metadata:

```go
meta := model.NewMeta(&model.MetaOptions{
    Definition: model.NewDefinition("A biological process...", []string{"GOC:go_curators"}),
    Synonyms:   []*model.Synonym{model.NewSynonym("hasExactSynonym", "biological process")},
    Comments:   []string{"Note: top-level class."},
    Version:    "2024-01-01",
})
```

## Storage

### DataSource Interface

The `storage.DataSource` interface abstracts persistent storage for OBO Graphs.

```go
type DataSource interface {
    SaveOboGraphInfo(grph graph.OboGraph) error
    UpdateOboGraphInfo(grph graph.OboGraph) error
    ExistsOboGraph(grph graph.OboGraph) bool
    SaveTerms(grph graph.OboGraph) (int, error)
    UpdateTerms(grph graph.OboGraph) (int, error)
    SaveOrUpdateTerms(grph graph.OboGraph) (*Stats, error)
    SaveRelationships(grph graph.OboGraph) (int, error)
    SaveNewRelationships(grph graph.OboGraph) (int, error)
}
```

### ArangoDB Backend

The `storage/arangodb` package implements `DataSource` against ArangoDB. It manages four collections:

| Collection | Stores |
|-----------|--------|
| Terms | Ontology terms (nodes) |
| Relationships | Term-to-term edges |
| Graph info | Ontology metadata (CV) |
| Named graph | Graph over terms and relationships |

Connect with explicit parameters or from an existing `arangomanager.Database`:

```go
import araobo "github.com/dictyBase/go-obograph/storage/arangodb"

// From connection parameters
ds, err := araobo.NewDataSource(
    &araobo.ConnectParams{
        User:     "root",
        Pass:     "secret",
        Database: "dictybase",
        Host:     "localhost",
        Port:     8529,
    },
    &araobo.CollectionParams{
        Term:         "cvterm",
        Relationship: "cvterm_relationship",
        GraphInfo:    "cv",
        OboGraph:     "obograph",
    },
)

// From an existing arangomanager.Database handle
ds, err := araobo.NewDataSourceFromDb(dbh, &araobo.CollectionParams{...})
```

| ConnectParams | Description |
|---------------|-------------|
| `User` | ArangoDB user |
| `Pass` | ArangoDB password |
| `Database` | Database name |
| `Host` | ArangoDB host |
| `Port` | ArangoDB port |
| `Istls` | Use TLS |

| CollectionParams | Description |
|------------------|-------------|
| `Term` | Collection for ontology terms |
| `Relationship` | Edge collection for relationships |
| `GraphInfo` | Collection for CV metadata |
| `OboGraph` | Named graph over terms and relationships |

### Load from JSON

`LoadOboJSONFromDataSource` reads an OBO JSON file, builds the graph in memory, and persists it via a `DataSource`.

```go
import "github.com/dictyBase/go-obograph/storage"

file, _ := os.Open("go.obo.json")
defer file.Close()

info, err := storage.LoadOboJSONFromDataSource(file, ds)
// info.IsCreated       — true if new, false if updated
// info.TermStats       — *Stats{Created, Updated, Deleted}
// info.RelationStats   — int (relationships created/updated)
```

## CLI Integration

The `command/flag` and `command/action` packages provide ready-to-use [urfave/cli](https://github.com/urfave/cli) integration for loading ontologies into ArangoDB.

### Flags

```go
import oboflag "github.com/dictyBase/go-obograph/command/flag"

app := cli.NewApp()
app.Flags = oboflag.OntologyFlags() // includes ArangoDB connection flags + --obojson
```

| Flag | Default | Description |
|------|---------|-------------|
| `--obojson, -j` | *(required)* | Input OBO JSON file(s), repeatable |
| `--term-collection` | `cvterm` | ArangoDB collection for ontology terms |
| `--rel-collection` | `cvterm_relationship` | Edge collection for relationships |
| `--cv-collection` | `cv` | Collection for ontology CV metadata |
| `--obograph` | `obograph` | Named graph for ontology traversal |

Plus ArangoDB connection flags from `arangomanager`: `--arangodb-user`, `--arangodb-pass`, `--arangodb-database`, `--arangodb-host`, `--arangodb-port`, `--is-secure`.

### CLI Action

Wire the `LoadOntologies` action into a urfave/cli application:

```go
import (
    oboaction "github.com/dictyBase/go-obograph/command/action"
    oboflag   "github.com/dictyBase/go-obograph/command/flag"
    "github.com/urfave/cli"
)

func main() {
    app := cli.NewApp()
    app.Name = "obo-loader"
    app.Flags = append([]cli.Flag{
        cli.StringFlag{Name: "log-format", Value: "json"},
        cli.StringFlag{Name: "log-level", Value: "error"},
    }, oboflag.OntologyFlags()...)
    app.Action = oboaction.LoadOntologies
    app.Run(os.Args)
}
```
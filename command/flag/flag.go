package flag

import (
	arangoflag "github.com/dictyBase/arangomanager/command/flag"
	"github.com/urfave/cli"
)

const (
	termCollectionFlag      = "term-collection"
	relCollectionFlag       = "rel-collection"
	cvCollectionFlag        = "cv-collection"
	obographFlag            = "obograph"
	cvtermValue             = "cvterm"
	cvtermRelationshipValue = "cvterm_relationship"
	cvValue                 = "cv"
	obographValue           = "obograph"
)

// OntologyFlagsOnly returns a slice of cli.Flag objects representing command line
// options for an ontology-related CLI application.
func OntologyFlagsOnly() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  termCollectionFlag,
			Usage: "arangodb collection for storing ontoloy terms",
			Value: cvtermValue,
		},
		cli.StringFlag{
			Name:  relCollectionFlag,
			Usage: "arangodb collection for storing cvterm relationships",
			Value: cvtermRelationshipValue,
		},
		cli.StringFlag{
			Name:  cvCollectionFlag,
			Usage: "arangodb collection for storing ontology information",
			Value: cvValue,
		},
		cli.StringFlag{
			Name:  obographFlag,
			Usage: "arangodb named graph for managing ontology graph",
			Value: obographValue,
		},
	}
}

// OntologyFlags returns a cli.flag slice to use in the command
// line arguments of the ontology loader.
func OntologyFlags() []cli.Flag {
	return append(
		[]cli.Flag{
			cli.StringFlag{
				Name:  termCollectionFlag,
				Usage: "arangodb collection for storing ontoloy terms",
				Value: cvtermValue,
			},
			cli.StringFlag{
				Name:  relCollectionFlag,
				Usage: "arangodb collection for storing cvterm relationships",
				Value: cvtermRelationshipValue,
			},
			cli.StringFlag{
				Name:  cvCollectionFlag,
				Usage: "arangodb collection for storing ontology information",
				Value: cvValue,
			},
			cli.StringFlag{
				Name:  obographFlag,
				Usage: "arangodb named graph for managing ontology graph",
				Value: obographValue,
			},
			cli.StringSliceFlag{
				Name:     "obojson,j",
				Usage:    "input ontology files in obograph json format",
				Required: true,
			},
		},
		arangoflag.ArangodbFlags()...,
	)
}

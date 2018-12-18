package flag

import (
	arangoflag "github.com/dictyBase/arangomanager/command/flag"
	cli "gopkg.in/urfave/cli.v1"
)

// OntologyFlags returns a cli.flag slice to use in the command
// line arguments of the ontology loader
func OntologyFlags() []cli.Flag {
	return append(
		[]cli.Flag{
			cli.StringFlag{
				Name:  "term-collection",
				Usage: "arangodb collection for storing ontoloy terms",
				Value: "cvterm",
			},
			cli.StringFlag{
				Name:  "rel-collection",
				Usage: "arangodb collection for storing cvterm relationships",
				Value: "cvterm_relationship",
			},
			cli.StringFlag{
				Name:  "cv-collection",
				Usage: "arangodb collection for storing ontology information",
				Value: "cv",
			},
			cli.StringFlag{
				Name:  "obograph",
				Usage: "arangodb named graph for managing ontology graph",
				Value: "obograph",
			},
			cli.StringSliceFlag{
				Name:  "obojson,j",
				Usage: "input ontology files in obograph json format",
			},
		},
		arangoflag.ArangodbFlags()...,
	)
}

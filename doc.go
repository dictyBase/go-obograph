/*
Package go-obograph is a golang library for handling OBO Graphs https://github.com/geneontology/obographs .
It provides API for the following...
  - Read JSON formatted OBO Graph file.
  - Build an in memory and read only graph structure for extracting information.
  - Persist the graph structure in arangodb database.

Example of a command line application to store OBO Graph in arangodb database

	import (
		"bufio"
		"context"
		"errors"
		"fmt"
		"log"
		"os"

		oboaction "github.com/dictyBase/go-obograph/command/action"
		oboflag "github.com/dictyBase/go-obograph/command/flag"
		"github.com/urfave/cli"
	)

	func main() {
		app := cli.NewApp()
		app.Name = "test cli for load obojson format file in arangodb"
		app.Flags = []cli.Flag{
			cli.StringFlag{
				Name:  "log-format",
				Usage: "format of the logging out, either of json or text.",
				Value: "json",
			},
			cli.StringFlag{
				Name:  "log-level",
				Usage: "log level for the application",
				Value: "error",
			},
			oboflag.OntologyFlags()...,
		}
		app.Action = oboaction.LoadOntologies
		if err := app.Run(os.Args); err != nil {
			log.Fatalf("error in running command %s", err)
		}
	}
*/package goobograph

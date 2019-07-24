package action

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dictyBase/go-obograph/graph"
	araobo "github.com/dictyBase/go-obograph/storage/arangodb"
	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

// LoadOntologies load ontologies into arangodb
func LoadOntologies(c *cli.Context) error {
	for _, v := range c.StringSlice("obojson") {
		r, err := os.Open(v)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in opening file %s %s", v, err),
				2,
			)
		}
		defer r.Close()
		g, err := graph.BuildGraph(r)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in building graph from %s %s", v, err),
				2,
			)
		}
		arPort, _ := strconv.Atoi(c.String("arangodb-port"))
		cp := &araobo.ConnectParams{
			User:     c.String("arangodb-user"),
			Pass:     c.String("arangodb-pass"),
			Host:     c.String("arangodb-host"),
			Database: c.String("arangodb-database"),
			Port:     arPort,
			Istls:    c.Bool("is-secure"),
		}
		clp := &araobo.CollectionParams{
			Term:         c.String("term-collection"),
			Relationship: c.String("rel-collection"),
			GraphInfo:    c.String("cv-collection"),
			OboGraph:     c.String("obograph"),
		}
		ds, err := araobo.NewDataSource(cp, clp)
		if err != nil {
			return cli.NewExitError(err.Error(), 2)
		}
		logger := getLogger(c)
		if !ds.ExistsOboGraph(g) {
			logger.Infof("obograph %s does not exist, have to be loaded", v)
			err := ds.SaveOboGraphInfo(g)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("error in saving graph %s", err),
					2,
				)
			}
			nt, err := ds.SaveTerms(g)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("error in saving terms %s", err),
					2,
				)
			}
			logger.Infof("saved %d terms", nt)
			nr, err := ds.SaveRelationships(g)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("error in saving relationships %s", err),
					2,
				)
			}
			logger.Infof("saved %d relationships", nr)
		}
	}
	return nil
}

func getLogger(c *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch c.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := c.GlobalString("log-level")
	switch l {
	case "debug":
		log.Level = logrus.DebugLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	}
	return logrus.NewEntry(log)
}

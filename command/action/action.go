package action

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dictyBase/go-obograph/graph"
	"github.com/dictyBase/go-obograph/storage"
	araobo "github.com/dictyBase/go-obograph/storage/arangodb"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const exitCode = 2

func ConnectParams(clt *cli.Context) *araobo.ConnectParams {
	arPort, _ := strconv.Atoi(clt.String("arangodb-port"))

	return &araobo.ConnectParams{
		User:     clt.String("arangodb-user"),
		Pass:     clt.String("arangodb-pass"),
		Host:     clt.String("arangodb-host"),
		Database: clt.String("arangodb-database"),
		Port:     arPort,
		Istls:    clt.Bool("is-secure"),
	}
}

func CollectParams(clt *cli.Context) *araobo.CollectionParams {
	return &araobo.CollectionParams{
		Term:         clt.String("term-collection"),
		Relationship: clt.String("rel-collection"),
		GraphInfo:    clt.String("cv-collection"),
		OboGraph:     clt.String("obograph"),
	}
}

func saveNewGraph(dsa storage.DataSource, grph graph.OboGraph, logger *logrus.Entry) error {
	err := dsa.SaveOboGraphInfo(grph)
	if err != nil {
		return fmt.Errorf("error in saving graph %s", err)
	}
	nst, err := dsa.SaveTerms(grph)
	if err != nil {
		return fmt.Errorf("error in saving terms %s", err)
	}
	logger.Infof("saved %d terms", nst)
	nsr, err := dsa.SaveRelationships(grph)
	if err != nil {
		return fmt.Errorf("error in saving relationships %s", err)
	}
	logger.Infof("saved %d relationships", nsr)

	return nil
}

func saveExistentGraph(dsa storage.DataSource, grph graph.OboGraph, logger *logrus.Entry) error {
	if err := dsa.UpdateOboGraphInfo(grph); err != nil {
		return fmt.Errorf("error in updating graph information %s", err)
	}
	stats, err := dsa.SaveOrUpdateTerms(grph)
	if err != nil {
		return fmt.Errorf("error in updating terms %s", err)
	}
	logger.Infof(
		"saved::%d terms updated::%d terms obsoleted::%d terms",
		stats.Created, stats.Updated, stats.Deleted,
	)
	urs, err := dsa.SaveNewRelationships(grph)
	if err != nil {
		return fmt.Errorf("error in saving relationships %s", err)
	}
	logger.Infof("updated %d relationships", urs)

	return nil
}

// LoadOntologies load ontologies into arangodb.
func LoadOntologies(clt *cli.Context) error {
	dsa, err := araobo.NewDataSource(ConnectParams(clt), CollectParams(clt))
	if err != nil {
		return cli.NewExitError(err.Error(), exitCode)
	}
	logger := getLogger(clt)
	for _, objs := range clt.StringSlice("obojson") {
		rdr, err := os.Open(objs)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in opening file %s %s", objs, err),
				exitCode,
			)
		}
		defer rdr.Close()
		grph, err := graph.BuildGraph(rdr)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error in building graph from %s %s", objs, err),
				exitCode,
			)
		}
		if !dsa.ExistsOboGraph(grph) {
			logger.Infof("obograph %s does not exist, have to be loaded", objs)
			if err := saveNewGraph(dsa, grph, logger); err != nil {
				return cli.NewExitError(err.Error(), exitCode)
			}

			continue
		}
		logger.Infof("obograph %s exist, have to be updated", objs)
		if err := saveExistentGraph(dsa, grph, logger); err != nil {
			return cli.NewExitError(err.Error(), exitCode)
		}
	}

	return nil
}

func getLogger(clt *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch clt.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := clt.GlobalString("log-level")
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

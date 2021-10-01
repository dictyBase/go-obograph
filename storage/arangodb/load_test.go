package arangodb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dictyBase/arangomanager/testarango"
	"github.com/dictyBase/go-obograph/storage"
	"github.com/stretchr/testify/require"
)

func oboReader(assert *require.Assertions) *os.File {
	dir, err := os.Getwd()
	if err != nil {
		assert.NoErrorf(err, "unable to get current dir %s", err)
	}
	fh, err := os.Open(
		filepath.Join(
			filepath.Dir(dir), "testdata", "dicty_phenotypes.json",
		),
	)
	assert.NoErrorf(err, "unable to open file %s", err)
	return fh
}

func tearDown(assert *require.Assertions, ta *testarango.TestArango) {
	dbh, err := ta.DB(ta.Database)
	assert.NoErrorf(
		err,
		"expect no error from getting the database instance, received error %s", err,
	)
	if err := dbh.Drop(); err != nil {
		assert.NoErrorf(
			err,
			"expect no error from dropping the database, received error %s", err,
		)
	}
}

func setUp(t *testing.T) (*require.Assertions, *testarango.TestArango, storage.DataSource) {
	ta, err := testarango.NewTestArangoFromEnv(true)
	if err != nil {
		t.Fatalf("unable to construct new TestArango instance %s", err)
	}
	assert := require.New(t)
	ds, err := NewDataSource(
		&ConnectParams{
			User:     ta.User,
			Pass:     ta.Pass,
			Host:     ta.Host,
			Database: ta.Database,
			Port:     ta.Port,
			Istls:    ta.Istls,
		}, &CollectionParams{
			Term:         "cvterm",
			Relationship: "cvterm_relationship",
			GraphInfo:    "cv",
			OboGraph:     "obograph",
		})
	assert.NoErrorf(
		err,
		"expect no error in making an arangodb datasource, received %s", err,
	)
	return assert, ta, ds
}

func TestLoadOboJSONFromDataSource(t *testing.T) {
	assert, ta, ds := setUp(t)
	r := oboReader(assert)
	defer tearDown(assert, ta)
	defer r.Close()
	info, err := storage.LoadOboJSONFromDataSource(r, ds)
	assert.NoErrorf(err, "expect no error from loading, received %s", err)
	assert.True(info.IsCreated, "expect the obo data to be created")
	assert.Equal(info.RelationStats, 1143, "should load 1143 relationships")
	assert.Equal(info.TermStats.Created, 1271, "should load 1271 terms")
	assert.Equal(info.TermStats.Updated, 0, "should have not updated any term")
	assert.Equal(info.TermStats.Deleted, 0, "should have not deleted any term")
}

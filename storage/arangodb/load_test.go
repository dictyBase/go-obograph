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
	fho, err := os.Open(
		filepath.Join(
			filepath.Dir(dir), "testdata", "dicty_phenotypes.json",
		),
	)
	assert.NoErrorf(err, "unable to open file %s", err)

	return fho
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
	t.Helper()
	tra, err := testarango.NewTestArangoFromEnv(true)
	if err != nil {
		t.Fatalf("unable to construct new TestArango instance %s", err)
	}
	assert := require.New(t)
	dsr, err := NewDataSource(
		&ConnectParams{
			User:     tra.User,
			Pass:     tra.Pass,
			Host:     tra.Host,
			Database: tra.Database,
			Port:     tra.Port,
			Istls:    tra.Istls,
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

	return assert, tra, dsr
}

func TestLoadOboJSONFromDataSource(t *testing.T) {
	t.Parallel()
	assert, ta, dsr := setUp(t)
	r := oboReader(assert)
	defer tearDown(assert, ta)
	defer r.Close()
	info, err := storage.LoadOboJSONFromDataSource(r, dsr)
	assert.NoErrorf(err, "expect no error from loading, received %s", err)
	assert.True(info.IsCreated, "expect the obo data to be created")
	assert.Equal(info.RelationStats, 1143, "should load 1143 relationships")
	assert.Equal(info.TermStats.Created, 1271, "should load 1271 terms")
	assert.Equal(info.TermStats.Updated, 0, "should have not updated any term")
	assert.Equal(info.TermStats.Deleted, 0, "should have not deleted any term")
	r2 := oboReader(assert)
	info2, err := storage.LoadOboJSONFromDataSource(r2, dsr)
	assert.NoErrorf(err, "expect no error from reloading, received %s", err)
	assert.False(info2.IsCreated, "expect no obo data to be created")
	assert.Equal(info2.RelationStats, 0, "should not load any relationships")
	assert.Equal(info2.TermStats.Created, 0, "should not load any terms")
	assert.Equal(info2.TermStats.Updated, 1271, "should have updated 1271 term")
	assert.Equal(info2.TermStats.Deleted, 0, "should have not deleted any term")
}

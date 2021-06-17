package manager

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
)

// Database struct
type Database struct {
	dbh driver.Database
}

// Search query the database that is expected to return multiple rows of result
func (d *Database) Search(query string) (*Resultset, error) {
	// validate
	if err := d.dbh.ValidateQuery(context.Background(), query); err != nil {
		return &Resultset{empty: true}, fmt.Errorf("error in validating the query %s", err)
	}
	ctx := context.Background()
	c, err := d.dbh.Query(ctx, query, nil)
	if err != nil {
		if driver.IsNotFound(err) {
			return &Resultset{empty: true}, nil
		}
		return &Resultset{empty: true}, fmt.Errorf("error in doing query %s", err)
	}
	return &Resultset{cursor: c, ctx: ctx}, nil
}

// Count query the database that is expected to return count of result
func (d *Database) Count(query string) (int64, error) {
	// validate
	if err := d.dbh.ValidateQuery(context.Background(), query); err != nil {
		return 0, fmt.Errorf("error in validating the query %s", err)
	}
	c, err := d.dbh.Query(driver.WithQueryCount(context.Background(), true), query, nil)
	if err != nil {
		return 0, err
	}
	return c.Count(), nil
}

// Do is to run data modification query that is not expected to return any result
func (d *Database) Do(query string, bindVars map[string]interface{}) error {
	ctx := driver.WithSilent(context.Background())
	_, err := d.dbh.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("error in data modification query %s", err)
	}
	return nil
}

// Run is to run data modification query that is expected to return a result
// It is a convenient alias for Get method
func (d *Database) Run(query string) (*Result, error) {
	return d.Get(query)
}

// GetRow query the database with bind parameters that is expected to return
// single row of result
func (d *Database) GetRow(query string, bindVars map[string]interface{}) (*Result, error) {
	if err := d.dbh.ValidateQuery(context.Background(), query); err != nil {
		return &Result{empty: true}, fmt.Errorf("error in validating the query %s", err)
	}
	c, err := d.dbh.Query(context.Background(), query, bindVars)
	return d.getResult(c, err)
}

// Get query the database to return single row of result
func (d *Database) Get(query string) (*Result, error) {
	return d.GetRow(query, nil)
}

// FindOrCreateCollection finds or creates a collection in the database. The
// method is expected to be called by the user who has privileges to create the
// collection
func (d *Database) FindOrCreateCollection(name string, opt *driver.CreateCollectionOptions) (driver.Collection, error) {
	var c driver.Collection
	ok, err := d.dbh.CollectionExists(context.Background(), name)
	if err != nil {
		return c, fmt.Errorf("unable to check for collection %s", name)
	}
	if ok {
		return d.dbh.Collection(context.Background(), name)
	}
	return d.dbh.CreateCollection(context.TODO(), name, opt)
}

// Collection returns collection attached to current database
func (d *Database) Collection(name string) (driver.Collection, error) {
	var c driver.Collection
	ok, err := d.dbh.CollectionExists(context.Background(), name)
	if err != nil {
		return c, fmt.Errorf("unable to check for collection %s", name)
	}
	if !ok {
		return c, fmt.Errorf("collection %s has to be created", name)
	}
	return d.dbh.Collection(context.Background(), name)
}

// CreateCollection creates a collection in the database
func (d *Database) CreateCollection(name string, opt *driver.CreateCollectionOptions) (driver.Collection, error) {
	var c driver.Collection
	ok, err := d.dbh.CollectionExists(context.Background(), name)
	if err != nil {
		return c, fmt.Errorf("error in collection lookup %s", err)
	}
	if ok {
		return c, fmt.Errorf("collection %s exists", name)
	}
	return d.dbh.CreateCollection(context.Background(), name, opt)
}

// FindOrCreateGraph finds or creates a named graph in the database
func (d *Database) FindOrCreateGraph(name string, defs []driver.EdgeDefinition) (driver.Graph, error) {
	var g driver.Graph
	ok, err := d.dbh.GraphExists(context.Background(), name)
	if err != nil {
		return g, fmt.Errorf("error in graph %s lookup %s", name, err)
	}
	if ok {
		return d.dbh.Graph(context.Background(), name)
	}
	return d.dbh.CreateGraph(
		context.Background(),
		name,
		&driver.CreateGraphOptions{EdgeDefinitions: defs},
	)
}

// Drop removes the database
func (d *Database) Drop() error {
	return d.dbh.Remove(context.Background())
}

func (d *Database) getResult(c driver.Cursor, err error) (*Result, error) {
	if err != nil {
		return &Result{empty: true}, fmt.Errorf("error in query %s", err)
	}
	if !c.HasMore() {
		return &Result{empty: true}, nil
	}
	return &Result{cursor: c}, nil
}

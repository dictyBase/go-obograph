package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
)

func getDatabase(database string, client driver.Client) (driver.Database, error) {
	var db driver.Database
	ok, err := client.DatabaseExists(context.Background(), database)
	if err != nil {
		return db, fmt.Errorf("unable to check for database %s", err)
	}
	if !ok {
		return db, fmt.Errorf("database %s has to be created", database)
	}
	return client.Database(context.Background(), database)
}

func getCollection(db driver.Database, string collection) (driver.Collection, error) {
	var c driver.Collection
	ok, err := db.CollectionExists(context.Background(), collection)
	if err != nil {
		return c, fmt.Errorf("unable to check for collection %s", collection)
	}
	if !ok {
		return c, fmt.Errorf("collection %s has to be created", collection)
	}
	return db.Collection(context.Background(), collection)
}

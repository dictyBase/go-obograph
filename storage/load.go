package storage

import (
	"fmt"
	"io"

	"github.com/dictyBase/go-obograph/graph"
)

// UploadInformation gives information about obo upload
type UploadInformation struct {
	// IsCreated indicates whether the obo information is created or updated
	IsCreated bool
	// RelationStats gives no of relationships that are created
	RelationStats int
	// TermStats gives information about uploaded terms
	TermStats *Stats
}

// LoadOboJSONFromDataSource loads obojson from a given reader and datasource for storage
func LoadOboJSONFromDataSource(r io.Reader, ds DataSource) (*UploadInformation, error) {
	info := &UploadInformation{}
	g, err := graph.BuildGraph(r)
	if err != nil {
		return info, err
	}
	if ds.ExistsOboGraph(g) {
		return persistExistOboGraph(ds, g)
	}
	return persistNewOboGraph(ds, g)
}

func persistExistOboGraph(ds DataSource, g graph.OboGraph) (*UploadInformation, error) {
	info := &UploadInformation{IsCreated: false}
	if err := ds.UpdateOboGraphInfo(g); err != nil {
		return info, fmt.Errorf("error in updating graph information %s", err)
	}
	ts, err := ds.SaveOrUpdateTerms(g)
	if err != nil {
		return info, fmt.Errorf("error in updating terms %s", err)
	}
	rn, err := ds.SaveNewRelationships(g)
	if err != nil {
		return info, fmt.Errorf("error in saving relationships %s", err)
	}
	info.TermStats = ts
	info.RelationStats = rn
	return info, nil
}

func persistNewOboGraph(ds DataSource, g graph.OboGraph) (*UploadInformation, error) {
	info := &UploadInformation{IsCreated: true}
	if err := ds.SaveOboGraphInfo(g); err != nil {
		return info, fmt.Errorf("error in saving graph information %s", err)
	}
	tn, err := ds.SaveTerms(g)
	if err != nil {
		return info, fmt.Errorf("error in saving terms %s", err)
	}
	rn, err := ds.SaveRelationships(g)
	if err != nil {
		return info, fmt.Errorf("error in saving relationships %s", err)
	}
	info.TermStats = &Stats{Created: tn}
	info.RelationStats = rn
	return info, nil
}

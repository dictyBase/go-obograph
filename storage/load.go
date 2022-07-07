package storage

import (
	"fmt"
	"io"

	"github.com/dictyBase/go-obograph/graph"
)

// UploadInformation gives information about obo upload.
type UploadInformation struct {
	// IsCreated indicates whether the obo information is created or updated
	IsCreated bool
	// RelationStats gives no of relationships that are created
	RelationStats int
	// TermStats gives information about uploaded terms
	TermStats *Stats
}

// LoadOboJSONFromDataSource loads obojson from a given reader and datasource for storage.
func LoadOboJSONFromDataSource(r io.Reader, dsr DataSource) (*UploadInformation, error) {
	info := &UploadInformation{}
	grph, err := graph.BuildGraph(r)
	if err != nil {
		return info, fmt.Errorf("error in building graph %s", err)
	}
	if dsr.ExistsOboGraph(grph) {
		return persistExistOboGraph(dsr, grph)
	}

	return persistNewOboGraph(dsr, grph)
}

func persistExistOboGraph(dsr DataSource, grph graph.OboGraph) (*UploadInformation, error) {
	info := &UploadInformation{IsCreated: false}
	if err := dsr.UpdateOboGraphInfo(grph); err != nil {
		return info, fmt.Errorf("error in updating graph information %s", err)
	}
	trs, err := dsr.SaveOrUpdateTerms(grph)
	if err != nil {
		return info, fmt.Errorf("error in updating terms %s", err)
	}
	rn, err := dsr.SaveNewRelationships(grph)
	if err != nil {
		return info, fmt.Errorf("error in saving relationships %s", err)
	}
	info.TermStats = trs
	info.RelationStats = rn

	return info, nil
}

func persistNewOboGraph(dsr DataSource, grph graph.OboGraph) (*UploadInformation, error) {
	info := &UploadInformation{IsCreated: true}
	if err := dsr.SaveOboGraphInfo(grph); err != nil {
		return info, fmt.Errorf("error in saving graph information %s", err)
	}
	trm, err := dsr.SaveTerms(grph)
	if err != nil {
		return info, fmt.Errorf("error in saving terms %s", err)
	}
	rn, err := dsr.SaveRelationships(grph)
	if err != nil {
		return info, fmt.Errorf("error in saving relationships %s", err)
	}
	info.TermStats = &Stats{Created: trm}
	info.RelationStats = rn

	return info, nil
}

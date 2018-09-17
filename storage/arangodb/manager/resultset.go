package manager

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/fatih/structs"
)

// Resultset is a cursor for multiple rows of result
type Resultset struct {
	cursor driver.Cursor
	ctx    context.Context
	empty  bool
}

// IsEmpty checks for empty resultset
func (r *Resultset) IsEmpty() bool {
	return r.empty
}

// Scan advances resultset to the next row of data
func (r *Resultset) Scan() bool {
	if r.empty {
		return r.empty
	}
	return r.cursor.HasMore()
}

// Read read the row of data to interface i
func (r *Resultset) Read(i interface{}) error {
	meta, err := r.cursor.ReadDocument(r.ctx, i)
	if err != nil {
		return fmt.Errorf("error in reading document %s", err)
	}
	s := structs.New(i)
	if f, ok := s.FieldOk("DocumentMeta"); ok {
		if f.IsEmbedded() {
			if err := f.Set(meta); err != nil {
				return fmt.Errorf("error in assigning DocumentMeta to the structure %s", err)
			}
		}
	}
	return nil
}

// Close closed the resultset
func (r *Resultset) Close() error {
	return r.cursor.Close()
}

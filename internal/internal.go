package internal

import (
	"strings"
)

// ExtractID extracts the last part of an URL primary to create an unique id
// from the IRI values of graph and nodes
func ExtractID(s string) string {
	parts := strings.Split(s, "/")
	l := parts[len(parts)-1]
	if strings.Contains(l, "#") {
		mparts := strings.Split(l, "#")
		return mparts[len(mparts)-1]
	}
	return l
}

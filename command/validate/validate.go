package validate

import (
	"fmt"

	"github.com/urfave/cli"
)

// OntologyArgs validates the command line arguments
func OntologyArgs(c *cli.Context) error {
	for _, p := range []string{
		"arangodb-pass",
		"arangodb-database",
		"arangodb-user",
		"arangodb-host",
	} {
		if len(c.String(p)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", p),
				2,
			)
		}
	}
	return nil
}

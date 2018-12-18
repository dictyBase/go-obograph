package validate

import (
	"fmt"

	cli "gopkg.in/urfave/cli.v1"
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

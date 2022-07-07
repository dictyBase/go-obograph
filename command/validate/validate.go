package validate

import (
	"fmt"

	"github.com/urfave/cli"
)

const exitCode = 2

// OntologyArgs validates the command line arguments.
func OntologyArgs(clt *cli.Context) error {
	for _, param := range []string{
		"arangodb-pass",
		"arangodb-database",
		"arangodb-user",
		"arangodb-host",
	} {
		if len(clt.String(param)) == 0 {
			return cli.NewExitError(
				fmt.Sprintf("argument %s is missing", param),
				exitCode,
			)
		}
	}

	return nil
}

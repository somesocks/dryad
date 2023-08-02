package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var secretsPathCommand = clib.NewCommand("path", "print the path to the secrets for the current package, if it exists").
	WithArg(
		clib.
			NewArg("path", "path to the stem base dir").
			WithAutoComplete(ArgAutoCompletePath),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var err error
		var path string

		if len(args) > 0 {
			path = args[0]
			path, err = filepath.Abs(path)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}

		// normalize the path to point to the closest secrets
		path, err = dryad.SecretsPath(path)
		if err != nil {
			log.Fatal(err)
		}

		// check if the secrets folder exists
		exists, err := dryad.SecretsExist(path)
		if err != nil {
			log.Fatal(err)
		}

		if exists {
			fmt.Println(path)
		}

		return 0
	})

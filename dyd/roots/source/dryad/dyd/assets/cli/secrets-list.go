package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var secretsListCommand = clib.NewCommand("list", "list the secret files in a stem/root").
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

		err = dryad.SecretsWalk(
			dryad.SecretsWalkArgs{
				BasePath: path,
				OnMatch: func(path string, info fs.FileInfo) error {
					fmt.Println(path)
					return nil
				},
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

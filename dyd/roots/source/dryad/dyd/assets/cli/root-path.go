package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
)

var rootPathCommand = clib.NewCommand("path", "return the base path of the current root").
	WithArg(
		clib.
			NewArg("path", "the path to start searching for a root at. defaults to current directory").
			AsOptional().
			WithAutoComplete(AutoCompletePath),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var path string = ""

		if len(args) > 0 {
			path = args[0]
		}

		path, err := dryad.RootPath(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)

		return 0
	})

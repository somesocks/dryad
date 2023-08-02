package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var rootLinkCommand = clib.NewCommand("link", "link a root as a dependency of the current root").
	WithArg(
		clib.
			NewArg("path", "path to the root you want to link as a dependency").
			WithAutoComplete(ArgAutoCompletePath),
	).
	WithArg(clib.NewArg("alias", "the alias to link the root under. if not specified, this defaults to the basename of the linked root").AsOptional()).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var rootPath, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		var path = args[0]
		var alias = ""
		if len(args) > 1 {
			alias = args[1]
		}

		err = dryad.RootLink(rootPath, path, alias)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

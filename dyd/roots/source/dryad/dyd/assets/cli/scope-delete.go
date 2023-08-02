package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var scopeDeleteCommand = clib.NewCommand("delete", "remove an existing scope directory from the garden").
	WithArg(
		clib.
			NewArg("name", "the name of the scope to delete").
			WithAutoComplete(ArgAutoCompleteScope),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var name string = args[0]

		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		err = dryad.ScopeDelete(path, name)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

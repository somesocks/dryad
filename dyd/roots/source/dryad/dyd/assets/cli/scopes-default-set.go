package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var scopesDefaultSetCommand = clib.NewCommand("set", "set a scope to be the default").
	WithArg(
		clib.
			NewArg("name", "the name of the scope to set as default").
			WithAutoComplete(ArgAutoCompleteScope),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var name string = args[0]

		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		err = dryad.ScopeSetDefault(path, name)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var gardenCreateCommand = clib.NewCommand("create", "create a garden").
	WithArg(
		clib.
			NewArg("path", "the target path at which to create the garden").
			AsOptional().
			WithAutoComplete(AutoCompletePath),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var path string
		var err error

		if len(args) > 0 {
			path = args[0]
		}

		err = dryad.GardenCreate(path)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

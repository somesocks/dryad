package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var rootMoveCommand = clib.NewCommand("move", "move a root to a new location and correct all references").
	WithArg(
		clib.
			NewArg("source", "path to the source root").
			WithAutoComplete(AutoCompletePath),
	).
	WithArg(
		clib.
			NewArg("destination", "destination path for the root").
			WithAutoComplete(AutoCompletePath),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var source string = args[0]
		var dest string = args[1]

		err := dryad.RootMove(source, dest)

		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

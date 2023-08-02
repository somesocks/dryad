package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var rootCopyCommand = clib.NewCommand("copy", "make a copy of the specified root at a new location").
	WithArg(
		clib.
			NewArg("source", "path to the source root").
			WithAutoComplete(ArgAutoCompletePath),
	).
	WithArg(
		clib.
			NewArg("destination", "destination path for the root copy").
			WithAutoComplete(ArgAutoCompletePath),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var source string = args[0]
		var dest string = args[1]

		err := dryad.RootCopy(source, dest)

		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

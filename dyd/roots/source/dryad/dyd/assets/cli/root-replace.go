package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var rootReplaceCommand = clib.NewCommand("replace", "replace all references to one root with references to another").
	WithArg(clib.NewArg("source", "path to the source root")).
	WithArg(clib.NewArg("replacement", "path to the replacement root")).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var source string = args[0]
		var dest string = args[1]

		err := dryad.RootReplace(source, dest)

		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

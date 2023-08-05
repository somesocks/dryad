package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var rootReplaceCommand = func() clib.Command {
	command := clib.NewCommand("replace", "replace all references to one root with references to another").
		WithArg(
			clib.
				NewArg("source", "path to the source root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("replacement", "path to the replacement root").
				WithAutoComplete(ArgAutoCompletePath),
		).
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

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

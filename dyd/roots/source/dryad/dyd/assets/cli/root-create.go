package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var rootCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create a new root at the target path").
		WithArg(
			clib.
				NewArg("path", "the path to create the new root at").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string = args[0]

			err := dryad.RootCreate(path)

			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"

	zlog "github.com/rs/zerolog/log"
)

var rootMoveCommand = func() clib.Command {
	command := clib.NewCommand("move", "move a root to a new location and correct all references").
		WithArg(
			clib.
				NewArg("source", "path to the source root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("destination", "destination path for the root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var source string = args[0]
			var dest string = args[1]

			err := dryad.RootMove(source, dest)

			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

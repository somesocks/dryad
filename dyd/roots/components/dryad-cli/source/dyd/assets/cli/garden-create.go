package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"

	zlog "github.com/rs/zerolog/log"
)

var gardenCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create a garden").
		WithArg(
			clib.
				NewArg("path", "the target path at which to create the garden").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
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
				zlog.Fatal().Err(err).Msg("error while creating garden")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var gardenPruneCommand = func() clib.Command {
	command := clib.
		NewCommand("prune", "clear all build artifacts out of the garden not actively linked to a sprout or a root").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}
			err = dryad.GardenPrune(
				path,
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while pruning garden")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

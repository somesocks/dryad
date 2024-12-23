package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var gardenWipeCommand = func() clib.Command {
	command := clib.
		NewCommand("wipe", "clear all build artifacts out of the garden").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}
			err = dryad.GardenWipe(
				path,
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while wiping garden")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

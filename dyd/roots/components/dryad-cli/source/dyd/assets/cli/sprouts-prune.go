package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var sproutsPruneCommand = func() clib.Command {
	command := clib.NewCommand("prune", "synchronize the sprouts dir structure with the roots dir").
		WithAction(func(req clib.ActionRequest) int {
			path, err := os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err = dryad.SproutsPrune(garden)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while pruning sprouts")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

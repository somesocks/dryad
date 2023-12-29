package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var sproutsPruneCommand = func() clib.Command {
	command := clib.NewCommand("prune", "synchronize the sprouts dir structure with the roots dir").
		WithAction(func(req clib.ActionRequest) int {
			path, err := os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}
			err = dryad.SproutsPrune(path)
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

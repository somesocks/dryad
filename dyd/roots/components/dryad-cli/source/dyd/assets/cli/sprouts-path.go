package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var sproutsPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the path of the sprouts dir").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			err, garden := dryad.Garden(path).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err, sprouts := garden.Sprouts().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving sprouts")
				return 1
			}

			fmt.Println(sprouts.BasePath)

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

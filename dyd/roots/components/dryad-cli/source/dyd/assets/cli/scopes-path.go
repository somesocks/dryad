package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the path of the scopes dir").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				return 1
			}

			path, err = dryad.ScopesPath(&garden)

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding scopes path")
				return 1
			}
			fmt.Println(path)

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

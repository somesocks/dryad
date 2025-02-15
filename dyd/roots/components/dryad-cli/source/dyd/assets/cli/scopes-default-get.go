package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultGetCommand = func() clib.Command {
	command := clib.NewCommand("get", "return the name of the default scope, if set").
		WithAction(func(req clib.ActionRequest) int {

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			unsafeGarden := dryad.Garden(path)
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				return 1
			}

			scopeName, err := dryad.ScopeGetDefault(garden)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding active scope")
				return 1
			}

			if scopeName != "" {
				fmt.Println(scopeName)
			}

			return 0
		})
	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create a new scope directory for the garden").
		WithArg(clib.NewArg("name", "the name of the new scope")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var name string = args[0]

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

			scopePath, err := dryad.ScopeCreate(garden, name)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while creating scope")
				return 1
			}

			fmt.Println(scopePath)

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

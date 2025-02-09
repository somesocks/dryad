package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultUnsetCommand = func() clib.Command {
	command := clib.NewCommand("unset", "remove the default scope setting").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding active directory")
				return 1
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				return 1
			}

			err = dryad.ScopeUnsetDefault(&garden)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while removing active scope")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

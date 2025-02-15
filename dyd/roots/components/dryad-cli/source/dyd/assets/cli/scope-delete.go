package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeDeleteCommand = func() clib.Command {
	command := clib.NewCommand("delete", "remove an existing scope directory from the garden").
		WithArg(
			clib.
				NewArg("name", "the name of the scope to delete").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var name string = args[0]

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				return 1
			}

			err = dryad.ScopeDelete(garden, name)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while deleting scope")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

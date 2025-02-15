package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create a new root at the target path").
		WithArg(
			clib.
				NewArg("path", "the path to create the new root at").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string = args[0]
			
			err, garden := dryad.Garden("").Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden roots")
				return 1
			}

			err, unsafeRoot := roots.Root(path).Clean()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving destination root location")
				return 1
			}
	
			err, safeRoot := unsafeRoot.Create(
				task.SERIAL_CONTEXT,
			)

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while creating root")
				return 1
			}

			fmt.Println(safeRoot.BasePath)
			return 0
		})

	command = LoggingCommand(command)


	return command
}()

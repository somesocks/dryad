package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var rootReplaceCommand = func() clib.Command {
	command := clib.NewCommand("replace", "replace all references to one root with references to another").
		WithArg(
			clib.
				NewArg("source", "path to the source root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("replacement", "path to the replacement root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var source string = args[0]
			var dest string = args[1]

			err, garden := dryad.Garden(source).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving garden")
				return 1
			}	

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden roots")
				return 1
			}

			err, safeSourceRoot := roots.Root(source).Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving source root")
				return 1
			}
	
			err, safeDestRoot := roots.Root(dest).Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving dest root")
				return 1
			}	

			err = dryad.RootReplace(
				dryad.RootReplaceRequest{
					Source: &safeSourceRoot,
					Dest: &safeDestRoot,
				},
			)

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while replacing root")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

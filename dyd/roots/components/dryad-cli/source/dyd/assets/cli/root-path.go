package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the base path of the current root").
		WithArg(
			clib.
				NewArg("path", "the path to start searching for a root at. defaults to current directory").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			err, path = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while cleaning root path")
				return 1
			}

			err, garden := dryad.Garden(path).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving garden")
				return 1
			}

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving roots")
				return 1
			}

			err, root := roots.Root(path).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root")
				return 1
			}

			fmt.Println(root.BasePath)

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementsAddCommand = func() clib.Command {
	command := clib.NewCommand("add", "add a root as a dependency of the current root").
		WithArg(
			clib.
				NewArg("path", "path to the root you want to add as a dependency").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(clib.NewArg("alias", "the alias to add the root under. if not specified, this defaults to the basename of the linked root").AsOptional()).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var rootPath, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			var depPath = args[0]
			var alias = ""
			if len(args) > 1 {
				alias = args[1]
			}

			err, rootPath = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, rootPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root path")
				return 1
			}

			err, depPath = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, depPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving dependency path")
				return 1
			}

			err, garden := dryad.Garden(rootPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving garden")
				return 1
			}

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving roots")
				return 1
			}

			err, root := roots.Root(rootPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root")
				return 1
			}

			err, dep := roots.Root(depPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving dependency")
				return 1
			}

			err = root.Link(
				task.SERIAL_CONTEXT,
				dryad.RootLinkRequest{
					Dependency: &dep,
					Alias: alias,
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while linking root")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

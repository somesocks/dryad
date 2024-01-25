package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
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

			var path = args[0]
			var alias = ""
			if len(args) > 1 {
				alias = args[1]
			}

			err = dryad.RootLink(rootPath, path, alias)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while linking root")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

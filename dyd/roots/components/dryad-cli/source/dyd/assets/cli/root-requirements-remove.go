package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementsRemoveCommand = func() clib.Command {
	command := clib.NewCommand("remove", "remove a requirement from the current root").
		WithArg(
			clib.
				NewArg("path", "path to the dependency to remove").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var rootPath, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			var path = args[0]

			err = dryad.RootUnlink(rootPath, path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while unlinking root")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

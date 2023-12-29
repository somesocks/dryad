package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootUnlinkCommand = func() clib.Command {
	command := clib.NewCommand("unlink", "remove a dependency linked to the current root").
		WithArg(
			clib.
				NewArg("path", "path to the dependency to unlink").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var rootPath, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			var path = args[0]

			err = dryad.RootUnlink(rootPath, path)
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

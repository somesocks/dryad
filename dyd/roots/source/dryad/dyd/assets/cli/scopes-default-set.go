package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultSetCommand = func() clib.Command {
	command := clib.NewCommand("set", "set a scope to be the default").
		WithArg(
			clib.
				NewArg("name", "the name of the scope to set as default").
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

			err = dryad.ScopeSetDefault(path, name)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while setting active scope")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

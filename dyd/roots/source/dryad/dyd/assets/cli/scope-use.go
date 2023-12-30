package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeUseCommand = func() clib.Command {
	command := clib.NewCommand("use", "set a scope to be active. alias for `dryad scopes default set`").
		WithArg(
			clib.
				NewArg("name", "the name of the scope to set as active. use 'none' to unset the active scope").
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

			if name == "none" {
				err = dryad.ScopeUnsetDefault(path)
			} else {
				err = dryad.ScopeSetDefault(path, name)
			}

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while setting active scope")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

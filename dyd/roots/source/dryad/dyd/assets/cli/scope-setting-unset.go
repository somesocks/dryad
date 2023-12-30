package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeSettingUnsetCommand = func() clib.Command {
	command := clib.NewCommand("unset", "remove a setting from a scope").
		WithArg(
			clib.
				NewArg("scope", "the name of the scope").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithArg(clib.NewArg("setting", "the name of the setting")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var scope string = args[0]
			var setting string = args[1]

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			err = dryad.ScopeSettingUnset(path, scope, setting)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while removing scope setting")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeSettingSetCommand = func() clib.Command {
	command := clib.NewCommand("set", "set the value of a setting in a scope").
		WithArg(
			clib.
				NewArg("scope", "the name of the scope").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithArg(clib.NewArg("setting", "the name of the setting")).
		WithArg(clib.NewArg("value", "the new value for the setting")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var scope string = args[0]
			var setting string = args[1]
			var value string = args[2]

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			unsafeGarden := dryad.Garden(path)
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				return 1
			}

			err = dryad.ScopeSettingSet(garden, scope, setting, value)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while changing scope setting")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

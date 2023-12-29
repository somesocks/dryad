package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeSettingGetCommand = func() clib.Command {
	command := clib.NewCommand("get", "print the value of a setting in a scope, if it exists").
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
				zlog.Fatal().Err(err)
				return 1
			}

			value, err := dryad.ScopeSettingGet(path, scope, setting)
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}

			if value != "" {
				fmt.Println(value)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

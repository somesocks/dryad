package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
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
				log.Fatal(err)
			}

			err = dryad.ScopeSettingSet(path, scope, setting, value)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var scopesDefaultUnsetCommand = func() clib.Command {
	command := clib.NewCommand("unset", "remove the default scope setting").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.ScopeUnsetDefault(path)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

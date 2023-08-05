package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var scopeActiveCommand = func() clib.Command {
	command := clib.NewCommand("active", "return the name of the active scope, if set. alias for `dryad scopes default get`").
		WithAction(func(req clib.ActionRequest) int {

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			scopeName, err := dryad.ScopeGetDefault(path)
			if err != nil {
				log.Fatal(err)
			}

			if scopeName != "" {
				fmt.Println(scopeName)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

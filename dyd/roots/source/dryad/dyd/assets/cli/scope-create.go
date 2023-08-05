package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var scopeCreateCommand = func() clib.Command {
	command := clib.NewCommand("create", "create a new scope directory for the garden").
		WithArg(clib.NewArg("name", "the name of the new scope")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var name string = args[0]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			scopePath, err := dryad.ScopeCreate(path, name)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(scopePath)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

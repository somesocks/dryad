package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var scopesPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the path of the scopes dir").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.ScopesPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

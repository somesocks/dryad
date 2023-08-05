package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
)

var gardenPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the base path for a garden").
		WithArg(
			clib.
				NewArg("path", "the target path at which to start for the base garden path").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			path, err = dryad.GardenPath(path)
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

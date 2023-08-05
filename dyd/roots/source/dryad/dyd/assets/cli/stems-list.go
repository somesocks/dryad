package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"log"
	"os"
)

var stemsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all stems that are dependencies for the current root").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.StemsWalk(path, func(path string, info fs.FileInfo, err error) error {
				fmt.Println(path)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

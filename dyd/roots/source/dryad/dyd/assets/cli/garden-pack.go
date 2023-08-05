package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
)

var gardenPackCommand = func() clib.Command {
	command := clib.NewCommand("pack", "pack the current garden into an archive ").
		WithArg(
			clib.
				NewArg("gardenPath", "the path to the garden to pack").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("targetPath", "the path (including name) to output the archive to").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var gardenPath = ""
			var targetPath = ""
			switch len(args) {
			case 0:
				break
			case 1:
				gardenPath = args[0]
			default:
				gardenPath = args[0]
				targetPath = args[1]
			}

			targetPath, err := dryad.GardenPack(gardenPath, targetPath)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(targetPath)
			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

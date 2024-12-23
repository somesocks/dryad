package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"

	zlog "github.com/rs/zerolog/log"
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
				zlog.Fatal().Err(err).Msg("error while packing garden")
				return 1
			}

			fmt.Println(targetPath)
			return 0
		})

	command = LoggingCommand(command)


	return command
}()

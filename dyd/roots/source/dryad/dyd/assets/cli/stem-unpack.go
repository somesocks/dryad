package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemUnpackCommand = func() clib.Command {
	command := clib.NewCommand("unpack", "unpack a stem archive at the target path and import it into the current garden").
		WithArg(
			clib.
				NewArg("archive", "the path to the archive to unpack").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var stemPath = args[0]

			gardenPath, err := os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			targetPath, err := dryad.StemUnpack(gardenPath, stemPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while unpacking stem")
				return 1
			}

			fmt.Println(targetPath)
			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

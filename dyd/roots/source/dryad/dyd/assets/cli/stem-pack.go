package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var stemPackCommand = func() clib.Command {
	command := clib.NewCommand("pack", "pack the stem at the target path into a tar archive").
		WithArg(
			clib.
				NewArg("stemPath", "the path to the stem to pack").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(clib.NewArg("targetPath", "the path (including name) to output the archive to").AsOptional()).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var stemPath = args[0]
			var targetPath = ""
			if len(args) > 1 {
				targetPath = args[1]
			}

			targetPath, err := dryad.StemPack(stemPath, targetPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while packing stem")
				return 1
			}

			fmt.Println(targetPath)
			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

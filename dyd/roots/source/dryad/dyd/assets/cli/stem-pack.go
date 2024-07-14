package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var stemPackCommand = func() clib.Command {
	command := clib.NewCommand("pack", "export the stem at the target path into a new folder or archive").
		WithArg(
			clib.
				NewArg("stemPath", "the path to the stem to pack").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("targetPath", "the path (including name) to output the archive to").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.
				NewOption("format", "export format. can be one of 'dir', 'tar', or 'tar.gz'. defaults to 'tar.gz'").
				WithType(clib.OptionTypeString),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts
			var format string

			var stemPath = args[0]
			var targetPath = args[1]

			if options["format"] != nil {
				format = options["format"].(string)
			} else {
				format = "tar.gz"
			}

			targetPath, err := dryad.StemPack(dryad.StemPackRequest{
				StemPath : stemPath,
				TargetPath: targetPath,	
				Format: format,			
			})
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

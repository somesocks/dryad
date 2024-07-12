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
				NewOption("includeDependencies", "include direct and indirect dependencies of the stem into the export. default false").
				WithType(clib.OptionTypeBool),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var stemPath = args[0]
			var targetPath = args[1]
			var includeDependencies bool

			if options["includeDependencies"] != nil {
				includeDependencies = options["includeDependencies"].(bool)
			} else {
				includeDependencies = false
			}

			targetPath, err := dryad.StemPack(dryad.StemPackRequest{
				StemPath : stemPath,
				TargetPath: targetPath,	
				IncludeDependencies: includeDependencies,			
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

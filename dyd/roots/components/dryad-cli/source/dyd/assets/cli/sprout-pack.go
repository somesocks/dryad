package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var sproutPackCommand = func() clib.Command {
	type ParsedArgs struct {
		SproutPath string
		TargetPath string
		Format     string
		Parallel   int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts
		var format string
		var parallel int

		sproutPath := args[0]
		targetPath := args[1]

		if options["format"] != nil {
			format = options["format"].(string)
		} else {
			format = "tar.gz"
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			SproutPath: sproutPath,
			TargetPath: targetPath,
			Format:     format,
			Parallel:   parallel,
		}
	}

	var runPack = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		targetPath, err := dryad.SproutPack(
			ctx,
			dryad.SproutPackRequest{
				SourceSproutPath: args.SproutPath,
				TargetPath:       args.TargetPath,
				Format:           args.Format,
			},
		)
		if err != nil {
			return err, nil
		}

		fmt.Println(targetPath)
		return nil, nil
	}

	runPack = task.WithContext(
		runPack,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runPack,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while packing sprout")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("pack", "export the sprout at the target path into a new folder or archive").
		WithArg(
			clib.
				NewArg("sproutPath", "the path to the sprout to pack").
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
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

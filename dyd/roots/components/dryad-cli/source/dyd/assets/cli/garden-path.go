package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var gardenPathCommand = func() clib.Command {
	type ParsedArgs struct {
		GardenPath string
		Parallel   int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var path string
		var parallel int

		if len(args) > 0 {
			path = args[0]
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			GardenPath: path,
			Parallel:   parallel,
		}
	}

	var printGardenPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		fmt.Println(garden.BasePath)
		return nil, nil
	}

	printGardenPath = task.WithContext(
		printGardenPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printGardenPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding garden path")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("path", "return the base path for a garden").
		WithArg(
			clib.
				NewArg("path", "the target path at which to start for the base garden path").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

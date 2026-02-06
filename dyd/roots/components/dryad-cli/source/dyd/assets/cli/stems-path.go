package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemsPathCommand = func() clib.Command {
	type ParsedArgs struct {
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var options = req.Opts
		var parallel int

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Parallel: parallel,
		}
	}

	var printStemsPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		path, err := os.Getwd()
		if err != nil {
			return err, nil
		}

		path, err = dryad.StemsPath(path)
		if err != nil {
			return err, nil
		}

		fmt.Println(path)
		return nil, nil
	}

	printStemsPath = task.WithContext(
		printStemsPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printStemsPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding stems path")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("path", "return the path of the stems dir").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

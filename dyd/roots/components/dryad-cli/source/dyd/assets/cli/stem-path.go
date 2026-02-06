package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemPathCommand = func() clib.Command {
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

	var printStemPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		path, err := os.Getwd()
		if err != nil {
			return err, nil
		}

		path, err = dryad.StemPath(path)
		if err != nil {
			return err, nil
		}

		fmt.Println(path)
		return nil, nil
	}

	printStemPath = task.WithContext(
		printStemPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printStemPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding stem path")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("path", "return the base path of the current root").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

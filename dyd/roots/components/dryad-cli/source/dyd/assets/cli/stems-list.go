package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"io/fs"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemsListCommand = func() clib.Command {
	type ParsedArgs struct {
		Path     string
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var options = req.Opts
		var parallel int

		path, err := os.Getwd()
		if err != nil {
			return err, ParsedArgs{}
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Path:     path,
			Parallel: parallel,
		}
	}

	var listStems = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err := dryad.StemsWalk(args.Path, func(path string, info fs.FileInfo, err error) error {
			fmt.Println(path)
			return nil
		})
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	listStems = task.WithContext(
		listStems,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			listStems,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling stems")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("list", "list all stems that are dependencies for the current root").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

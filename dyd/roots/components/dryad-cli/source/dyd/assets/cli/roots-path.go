package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootsPathCommand = func() clib.Command {
	type ParsedArgs struct {
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var options = req.Opts
			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			var path, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var printRootsPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		path, err := dryad.RootsPath(garden)
		if err != nil {
			return err, nil
		}
		fmt.Println(path)

		return nil, nil
	}

	printRootsPath = task.WithContext(
		printRootsPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printRootsPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding roots path")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("path", "return the path of the roots dir").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

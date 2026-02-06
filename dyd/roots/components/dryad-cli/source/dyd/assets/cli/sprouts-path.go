package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var sproutsPathAction = func(req clib.ActionRequest) int {
	type ParsedArgs struct {
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var opts = req.Opts
			var parallel int

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			path, err := os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var printSproutsPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.GardenPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, sprouts := garden.Sprouts().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		fmt.Println(sprouts.BasePath)
		return nil, nil
	}

	printSproutsPath = task.WithContext(
		printSproutsPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	return task.Return(
		task.Series2(
			parseArgs,
			printSproutsPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding sprouts path")
				return 1
			}

			return 0
		},
	)(req)
}

var sproutsPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the path of the sprouts dir").
		WithAction(sproutsPathAction)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

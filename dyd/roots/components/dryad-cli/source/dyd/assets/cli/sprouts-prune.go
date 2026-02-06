package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var sproutsPruneAction = func(req clib.ActionRequest) int {
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

	var pruneSprouts = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.GardenPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, sprouts := garden.Sprouts().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = sprouts.Prune(ctx)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	pruneSprouts = task.WithContext(
		pruneSprouts,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	return task.Return(
		task.Series2(
			parseArgs,
			pruneSprouts,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while pruning sprouts")
				return 1
			}

			return 0
		},
	)(req)
}

var sproutsPruneCommand = func() clib.Command {
	command := clib.NewCommand("prune", "synchronize the sprouts dir structure with the roots dir").
		WithAction(sproutsPruneAction)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

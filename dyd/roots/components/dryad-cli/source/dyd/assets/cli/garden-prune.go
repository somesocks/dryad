package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"time"

	zlog "github.com/rs/zerolog/log"
)

var gardenPruneCommand = func() clib.Command {

	type ParsedArgs struct {
		Path string
		Parallel int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var path string
			// var err error

			if len(args) > 0 {
				path = args[0]
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}
	
			return nil, ParsedArgs{
				Path: path,
				Parallel: parallel,
			}
		},
	)

	var pruneGarden = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.UnsafeGardenReference{
			BasePath: args.Path,
		}
		
		err, garden := unsafeGarden.Resolve(ctx, nil)
		if err != nil {
			return err, nil
		}

		err, _ = dryad.GardenPrune(
			ctx,
			dryad.GardenPruneRequest{
				Garden: garden,
				Snapshot: time.Now().Local(),
			},
		)
		return err, nil
	}

	pruneGarden = task.WithContext(
		pruneGarden,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			pruneGarden,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while pruning garden")
				return 1
			}

			return 0
		},
	)


	command := clib.
		NewCommand("prune", "clear all build artifacts out of the garden not actively linked to a sprout or a root").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)


	return command
}()


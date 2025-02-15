package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var gardenWipeCommand = func() clib.Command {

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

	var wipeGarden = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.UnsafeGardenReference{
			BasePath: args.Path,
		}
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = garden.Wipe(ctx) 

		return err, nil
	}

	wipeGarden = task.WithContext(
		wipeGarden,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			wipeGarden,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while wiping garden")
				return 1
			}

			return 0
		},
	)

	command := clib.
		NewCommand("wipe", "clear all build artifacts out of the garden").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	task "dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var gardenCreateCommand = func() clib.Command {

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

	var createGarden = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, _ := dryad.GardenCreate(ctx, dryad.GardenCreateRequest{BasePath: args.Path})
		return err, nil
	}

	createGarden = task.WithContext(
		createGarden,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.BuildContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			createGarden,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while creating garden")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("create", "create a garden").
		WithArg(
			clib.
				NewArg("path", "the target path at which to create the garden").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var rootMoveCommand = func() clib.Command {

	type ParsedArgs struct {
		SourcePath string
		DestPath string
		Parallel int
	}	

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var source string = args[0]
			var dest string = args[1]

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}
	
			return nil, ParsedArgs{
				SourcePath: source,
				DestPath: dest,
				Parallel: parallel,
			}
		},
	)

	var moveRoot = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.UnsafeGardenReference{
			BasePath: args.SourcePath,
		}
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		unsafeSourceRoot := dryad.UnsafeRootReference{
			BasePath: args.SourcePath,
			Garden: garden,
		}

		err, safeSourceRoot := unsafeSourceRoot.Resolve(ctx, nil)
		if err != nil {
			return err, nil
		}

		unsafeDestRoot := dryad.UnsafeRootReference{
			BasePath: args.DestPath,
			Garden: garden,
		}

		err, unsafeDestRoot = unsafeDestRoot.Clean()
		if err != nil {
			return err, nil
		}

		err, _ = dryad.RootMove(
			ctx,
			dryad.RootMoveRequest{
				Source: &safeSourceRoot,
				Dest: &unsafeDestRoot,
			},
		)
		return err, nil
	}

	moveRoot = task.WithContext(
		moveRoot,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			moveRoot,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error during root move")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("move", "move a root to a new location and correct all references").
		WithArg(
			clib.
				NewArg("source", "path to the source root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("destination", "destination path for the root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"

	zlog "github.com/rs/zerolog/log"
)

var rootMoveCommand = func() clib.Command {

	type ParsedArgs struct {
		SourcePath string
		DestPath string
		Parallel int
	}	

	var parseArgs =
		func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var source string = args[0]
			var dest string = args[1]
			var err error

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}

			err, source = dydfs.PartialEvalSymlinks(ctx, source)
			if err != nil {
				return err, ParsedArgs{}
			}			

			err, dest = dydfs.PartialEvalSymlinks(ctx, dest)
			if err != nil {
				return err, ParsedArgs{}
			}
			
			return nil, ParsedArgs{
				SourcePath: source,
				DestPath: dest,
				Parallel: parallel,
			}
		}

	var moveRoot = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		
		err, garden := dryad.Garden(args.SourcePath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeSourceRoot := roots.Root(args.SourcePath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		unsafeDestRoot := roots.Root(args.DestPath)

		err = safeSourceRoot.Move(
			ctx,
			dryad.RootMoveRequest{
				Dest: unsafeDestRoot,
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

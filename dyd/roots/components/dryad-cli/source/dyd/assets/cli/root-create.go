package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"

	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootCreateCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath string
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = 
		func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts
			var err error 

			var rootPath string

			if len(args) > 0 {
				rootPath = args[0]
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}

			rootPath, err = filepath.Abs(rootPath)
			if err != nil {
				return err, ParsedArgs{}
			}
	
			return nil, ParsedArgs{
				RootPath: rootPath,
				Parallel: parallel,
			}
		}

	var createRoot task.Task[ParsedArgs, any] =
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {

			err, garden := dryad.Garden("").Resolve(ctx)
			if err != nil {
				return err, nil
			}

			err, roots := garden.Roots().Resolve(ctx)
			if err != nil {
				return err, nil
			}

			err, unsafeRoot := roots.Root(args.RootPath).Clean()
			if err != nil {
				return err, nil
			}
	
			err, safeRoot := unsafeRoot.Create(
				ctx,
			)

			fmt.Println(safeRoot.BasePath)

			return nil, nil
		}

	createRoot = task.WithContext(
		createRoot,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			createRoot,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while creating root")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("create", "create a new root at the target path").
		WithArg(
			clib.
				NewArg("path", "the path to create the new root at").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootPathCommand = func() clib.Command {
	type ParsedArgs struct {
		Path     string
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var path string

		if len(args) > 0 {
			path = args[0]
		}

		var parallel int

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, path := dydfs.PartialEvalSymlinks(ctx, path)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			Path:     path,
			Parallel: parallel,
		}
	}

	var printRootPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		path := args.Path

		err, garden := dryad.Garden(path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, root := roots.Root(path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		fmt.Println(root.BasePath)
		return nil, nil
	}

	printRootPath = task.WithContext(
		printRootPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printRootPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root path")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("path", "return the base path of the current root").
		WithArg(
			clib.
				NewArg("path", "the path to start searching for a root at. defaults to current directory").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

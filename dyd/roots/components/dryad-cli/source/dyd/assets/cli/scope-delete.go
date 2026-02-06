package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeDeleteCommand = func() clib.Command {
	type ParsedArgs struct {
		Name       string
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var opts = req.Opts

			var name string = args[0]
			var parallel int

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			var path, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				Name:       name,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var deleteScope = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = dryad.ScopeDelete(garden, args.Name)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	deleteScope = task.WithContext(
		deleteScope,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			deleteScope,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while deleting scope")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("delete", "remove an existing scope directory from the garden").
		WithArg(
			clib.
				NewArg("name", "the name of the scope to delete").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

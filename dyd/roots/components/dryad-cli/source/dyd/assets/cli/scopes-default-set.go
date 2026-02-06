package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultSetCommand = func() clib.Command {
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

	var setDefaultScope = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = dryad.ScopeSetDefault(garden, args.Name)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	setDefaultScope = task.WithContext(
		setDefaultScope,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			setDefaultScope,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while setting active scope")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("set", "set a scope to be the default").
		WithArg(
			clib.
				NewArg("name", "the name of the scope to set as default").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

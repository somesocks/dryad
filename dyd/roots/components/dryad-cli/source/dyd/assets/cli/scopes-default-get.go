package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopesDefaultGetCommand = func() clib.Command {
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

			var path, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var getDefaultScope = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		scopeName, err := dryad.ScopeGetDefault(garden)
		if err != nil {
			return err, nil
		}

		if scopeName != "" {
			fmt.Println(scopeName)
		}

		return nil, nil
	}

	getDefaultScope = task.WithContext(
		getDefaultScope,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			getDefaultScope,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding active scope")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("get", "return the name of the default scope, if set").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

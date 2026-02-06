package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeSettingGetCommand = func() clib.Command {
	type ParsedArgs struct {
		Scope      string
		Setting    string
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var opts = req.Opts

			var scope string = args[0]
			var setting string = args[1]
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
				Scope:      scope,
				Setting:    setting,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var getScopeSetting = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		value, err := dryad.ScopeSettingGet(garden, args.Scope, args.Setting)
		if err != nil {
			return err, nil
		}

		if value != "" {
			fmt.Println(value)
		}

		return nil, nil
	}

	getScopeSetting = task.WithContext(
		getScopeSetting,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			getScopeSetting,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while getting scope setting")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("get", "print the value of a setting in a scope, if it exists").
		WithArg(
			clib.
				NewArg("scope", "the name of the scope").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithArg(clib.NewArg("setting", "the name of the setting")).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

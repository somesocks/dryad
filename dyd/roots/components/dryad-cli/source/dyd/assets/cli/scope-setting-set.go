package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeSettingSetCommand = func() clib.Command {
	type ParsedArgs struct {
		Scope      string
		Setting    string
		Value      string
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var opts = req.Opts

			var scope string = args[0]
			var setting string = args[1]
			var value string = args[2]
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
				Value:      value,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var setScopeSetting = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = dryad.ScopeSettingSet(garden, args.Scope, args.Setting, args.Value)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	setScopeSetting = task.WithContext(
		setScopeSetting,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			setScopeSetting,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while changing scope setting")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("set", "set the value of a setting in a scope").
		WithArg(
			clib.
				NewArg("scope", "the name of the scope").
				WithAutoComplete(ArgAutoCompleteScope),
		).
		WithArg(clib.NewArg("setting", "the name of the setting")).
		WithArg(clib.NewArg("value", "the new value for the setting")).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

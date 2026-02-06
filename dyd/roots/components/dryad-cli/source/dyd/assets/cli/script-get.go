package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"errors"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scriptGetAction = func(req clib.ActionRequest) int {
	type ParsedArgs struct {
		Command    string
		Scope      string
		HasScope   bool
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var opts = req.Opts
			var parallel int
			var scope string
			var hasScope bool

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			if opts["scope"] != nil {
				scope = opts["scope"].(string)
				hasScope = true
			}

			path, err := os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				Command:    args[0],
				Scope:      scope,
				HasScope:   hasScope,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var printScript = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		scope := args.Scope
		if !args.HasScope {
			scope, err = dryad.ScopeGetDefault(garden)
			zlog.Debug().Msg("loading default scope: " + scope)
			if err != nil {
				return err, nil
			}
		}

		if scope == "" || scope == "none" {
			return errors.New("no scope set, can't find command"), nil
		}
		zlog.Debug().Msg("using scope: " + scope)

		script, err := dryad.ScriptGet(dryad.ScriptGetRequest{
			Garden:  garden,
			Scope:   scope,
			Setting: "script-run-" + args.Command,
		})
		if err != nil {
			return err, nil
		}

		fmt.Println(script)

		return nil, nil
	}

	printScript = task.WithContext(
		printScript,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	return task.Return(
		task.Series2(
			parseArgs,
			printScript,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding script")
				return 1
			}
			return 0
		},
	)(req)
}

var scriptGetCommand = func() clib.Command {
	command := clib.NewCommand("get", "print the contents of a script").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithAction(scriptGetAction)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

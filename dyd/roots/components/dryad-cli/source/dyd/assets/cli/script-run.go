package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"errors"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var scriptRunAction = func(req clib.ActionRequest) int {
	type ParsedArgs struct {
		Command    string
		Args       []string
		Inherit    bool
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
			var inherit bool
			var scope string
			var hasScope bool

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			if opts["inherit"] != nil {
				inherit = opts["inherit"].(bool)
			} else {
				inherit = true
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
				Args:       args[1:],
				Inherit:    inherit,
				Scope:      scope,
				HasScope:   hasScope,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var runScript = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
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

		var env = map[string]string{}

		// setting scope to pass to shell
		env["DYD_SCOPE"] = scope

		// pull environment variables from parent process
		if args.Inherit {
			for _, e := range os.Environ() {
				if i := strings.Index(e, "="); i >= 0 {
					env[e[:i]] = e[i+1:]
				}
			}
		} else {
			// copy a few variables over from parent env for convenience
			env["TERM"] = os.Getenv("TERM")
		}

		err = dryad.ScriptRun(dryad.ScriptRunRequest{
			Garden:  garden,
			Scope:   scope,
			Setting: "script-run-" + args.Command,
			Args:    args.Args,
			Env:     env,
		})
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	runScript = task.WithContext(
		runScript,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	return task.Return(
		task.Series2(
			parseArgs,
			runScript,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while running script")
				return 1
			}
			return 0
		},
	)(req)
}

var scriptRunCommand = func() clib.Command {
	command := clib.NewCommand("run", "run a script in the current scope").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the alias to exec").WithType(clib.OptionTypeBool)).
		WithArg(clib.NewArg("-- args", "args to pass to the script").AsOptional()).
		WithAction(scriptRunAction)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

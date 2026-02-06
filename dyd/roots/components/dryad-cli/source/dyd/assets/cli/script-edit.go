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

var scriptEditAction = func(req clib.ActionRequest) int {
	type ParsedArgs struct {
		Command    string
		Editor     string
		HasEditor  bool
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
			var editor string
			var hasEditor bool
			var scope string
			var hasScope bool

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			if opts["editor"] != nil {
				editor = opts["editor"].(string)
				hasEditor = true
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
				Editor:     editor,
				HasEditor:  hasEditor,
				Scope:      scope,
				HasScope:   hasScope,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var editScript = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
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

		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				env[e[:i]] = e[i+1:]
			}
		}

		if args.HasEditor {
			env["EDITOR"] = args.Editor
		}

		err = dryad.ScriptEdit(dryad.ScriptEditRequest{
			Garden:  garden,
			Scope:   scope,
			Setting: "script-run-" + args.Command,
			Env:     env,
		})
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	editScript = task.WithContext(
		editScript,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	return task.Return(
		task.Series2(
			parseArgs,
			editScript,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while editing script")
				return 1
			}
			return 0
		},
	)(req)
}

var scriptEditCommand = func() clib.Command {
	command := clib.NewCommand("edit", "edit a script").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("editor", "set the editor to use")).
		WithAction(scriptEditAction)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopStartCommand = func() clib.Command {
	type ParsedArgs struct {
		Path       string
		Editor     string
		EditorArgs []string
		Inherit    bool
		OnExit     string
		Parallel   int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var opts = req.Opts

		var path string
		var editor string
		var editorArgs []string
		var inherit bool
		var onExit string
		var parallel int

		if len(args) > 0 {
			path = args[0]
		}

		if opts["editor"] != nil {
			editor = opts["editor"].(string)
		}

		if opts["arg"] != nil {
			editorArgs = opts["arg"].([]string)
		}

		if opts["on-exit"] != nil {
			onExit = opts["on-exit"].(string)
		}

		if opts["inherit"] != nil {
			inherit = opts["inherit"].(bool)
		}

		if opts["parallel"] != nil {
			parallel = int(opts["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Path:       path,
			Editor:     editor,
			EditorArgs: editorArgs,
			Inherit:    inherit,
			OnExit:     onExit,
			Parallel:   parallel,
		}
	}

	var runStart = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, path := dydfs.PartialEvalSymlinks(ctx, args.Path)
		if err != nil {
			return err, nil
		}

		err, garden := dryad.Garden(path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeRootRef := roots.Root(path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, rootFingerprint := safeRootRef.Develop(
			ctx,
			dryad.RootDevelopRequest{
				Editor:     args.Editor,
				EditorArgs: args.EditorArgs,
				Inherit:    args.Inherit,
				OnExit:     args.OnExit,
			},
		)
		if err != nil {
			return err, nil
		}
		fmt.Println(rootFingerprint)

		return nil, nil
	}

	runStart = task.WithContext(
		runStart,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runStart,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error from root development environment")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("start", "start a temporary development environment for a root").
		WithArg(
			clib.
				NewArg("path", "path to the root to develop").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("editor", "choose the editor to run in the root development environment").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("arg", "argument to pass to the editor").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("on-exit", "action to take when exiting with unsaved changes: ask, save, discard").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("inherit", "inherit env variables from the host environment").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ScopedCommand(command)
	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

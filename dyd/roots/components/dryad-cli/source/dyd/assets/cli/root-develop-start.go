package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopStartCommand = func() clib.Command {
	type ParsedArgs struct {
		Parallel  int
		Path      string
		Shell     string
		ShellArgs []string
		Inherit   bool
		OnExit    string
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var opts = req.Opts

		var path string
		var parallel int
		var shell string
		var shellArgs []string
		var inherit bool
		var onExit string

		if len(args) > 0 {
			if strings.HasPrefix(args[0], "-") {
				shellArgs = args
			} else {
				path = args[0]
				shellArgs = args[1:]
			}
		}

		if opts["parallel"] != nil {
			parallel = int(opts["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		if opts["shell"] != nil {
			shell = opts["shell"].(string)
		}

		if opts["on-exit"] != nil {
			onExit = opts["on-exit"].(string)
		}

		if opts["inherit"] != nil {
			inherit = opts["inherit"].(bool)
		}

		return nil, ParsedArgs{
			Parallel:  parallel,
			Path:      path,
			Shell:     shell,
			ShellArgs: shellArgs,
			Inherit:   inherit,
			OnExit:    onExit,
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
				Shell:     args.Shell,
				ShellArgs: args.ShellArgs,
				Inherit:   args.Inherit,
				OnExit:    args.OnExit,
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
		WithOption(clib.NewOption("shell", "choose the shell command to run in the root development environment").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("on-exit", "action to take when exiting with unsaved changes: ask, save, discard").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("inherit", "inherit env variables from the host environment").WithType(clib.OptionTypeBool)).
		WithArg(clib.NewArg("-- args", "args to pass to the shell").AsOptional()).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

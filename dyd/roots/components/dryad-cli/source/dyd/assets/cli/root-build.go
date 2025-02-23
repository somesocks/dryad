package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootBuildCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath string
		Parallel int
		JoinStdout bool
		JoinStderr bool
	}

	var parseArgs = 
		func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var path string
			var parallel int
			var joinStdout bool
			var joinStderr bool

			var err error

			if len(args) > 0 {
				path = args[0]
			}

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}

			if options["join-stdout"] != nil {
				joinStdout = options["join-stdout"].(bool)
			} else {
				joinStdout = false
			}

			if options["join-stderr"] != nil {
				joinStderr = options["join-stderr"].(bool)
			} else {
				joinStderr = false
			}

			err, path = dydfs.PartialEvalSymlinks(ctx, path)
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				RootPath: path,
				Parallel: parallel,
				JoinStdout: joinStdout,
				JoinStderr: joinStderr,
			}
		}

	var buildRoot = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		
		err, garden := dryad.Garden(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeRootRef := roots.Root(args.RootPath).Resolve(ctx, nil)
		if err != nil {
			return err, nil
		}

		var rootFingerprint string
		err, rootFingerprint = safeRootRef.Build(
			ctx,
			dryad.RootBuildRequest{
				JoinStdout: args.JoinStdout,
				JoinStderr: args.JoinStderr,
			},
		)
		if err != nil {
			return err, nil
		}
		fmt.Println(rootFingerprint)

		return nil, nil
	}

	buildRoot = task.WithContext(
		buildRoot,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			buildRoot,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building root")
				return 1
			}

			return 0
		},
	)

	command := clib.
		NewCommand("build", "build a specified root").
		WithArg(
			clib.
				NewArg("path", "path to the root to build").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.NewOption(
				"join-stdout",
				"join the stdout of child processes to the stderr of the parent dryad process. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.NewOption(
				"join-stderr",
				"join the stderr of child processes to the stderr of the parent dryad process. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

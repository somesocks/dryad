package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootBuildCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath          string
		VariantDescriptor string
		Parallel          int
		JoinStdout        bool
		JoinStderr        bool
		LogStdout         string
		LogStderr         string
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var rootRefRaw string
		var rootPath string
		var parallel int
		var variantDescriptor string
		var hasSelector bool
		var joinStdout bool
		var joinStderr bool
		var logStdout string
		var logStderr string

		var err error

		if len(args) > 0 {
			rootRefRaw = args[0]
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, rootRef := parseRootRef(rootRefRaw)
		if err != nil {
			return err, ParsedArgs{}
		}
		rootPath = rootRef.Path
		hasSelector = rootRef.HasSelector

		if hasSelector {
			err, variantDescriptor = (dryad.RootVariantContext{Descriptor: rootRef.Selector}).Filesystem()
			if err != nil {
				return err, ParsedArgs{}
			}
		}

		if options["variant"] != nil {
			if hasSelector {
				return fmt.Errorf("root build selector specified in both root_ref and --variant"), ParsedArgs{}
			}

			variantDescriptor = options["variant"].(string)
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

		if options["log-stdout"] != nil {
			logStdout = options["log-stdout"].(string)
			joinStdout = false
		}

		if options["log-stderr"] != nil {
			logStderr = options["log-stderr"].(string)
			joinStderr = false
		}

		err, rootPath = dydfs.PartialEvalSymlinks(ctx, rootPath)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			RootPath:          rootPath,
			VariantDescriptor: variantDescriptor,
			Parallel:          parallel,
			JoinStdout:        joinStdout,
			JoinStderr:        joinStderr,
			LogStdout:         logStdout,
			LogStderr:         logStderr,
		}
	}

	var buildRoot = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {

		err, garden := dryad.Garden(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeRootRef := roots.Root(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, sproutFingerprint := safeRootRef.BuildSprout(
			ctx,
			dryad.RootBuildSproutRequest{
				VariantDescriptor: args.VariantDescriptor,
				JoinStdout:        args.JoinStdout,
				JoinStderr:        args.JoinStderr,
				LogStdout: struct {
					Path string
					Name string
				}{
					Path: args.LogStdout,
					Name: "",
				},
				LogStderr: struct {
					Path string
					Name string
				}{
					Path: args.LogStderr,
					Name: "",
				},
			},
		)
		if err != nil {
			return err, nil
		}
		fmt.Println(sproutFingerprint)

		return nil, nil
	}

	buildRoot = task.WithContext(
		buildRoot,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			buildRoot,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while building root")
				return 1
			}

			return 0
		},
	)

	command := clib.
		NewCommand("build", "build a specified root").
		WithArg(
			clib.
				NewArg("root_ref", "path to the root to build, optionally qualified with a variant selector").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.NewOption(
				"variant",
				"select a root variant descriptor to build (filesystem form: dimension=option+dimension=option). supports none/any/host; inherit is invalid. defaults to all enabled variants",
			).
				WithType(clib.OptionTypeString),
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
		WithOption(clib.NewOption("log-stdout", "log the stdout of child processes to a file in the specified directory. disables joining").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("log-stderr", "log the stderr of child processes to a file in the specified directory. disables joining").WithType(clib.OptionTypeString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

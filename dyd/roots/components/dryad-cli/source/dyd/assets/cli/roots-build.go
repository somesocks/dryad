package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var rootsBuildCommand = func() clib.Command {

	type ParsedArgs struct {
		Filter            dryad.RootVariantFilter
		VariantDescriptor string
		Parallel          int
		Path              string
		JoinStdout        bool
		JoinStderr        bool
		LogStdout         string
		LogStderr         string
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		// var args = req.Args
		var options = req.Opts

		var path string
		// var err error

		var parallel int

		var variantDescriptor string
		var joinStdout bool
		var joinStderr bool
		var logStdout string
		var logStderr string

		if options["path"] != nil {
			path = options["path"].(string)
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		if options["variant"] != nil {
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

		err, rootFilter := ArgRootVariantFilterFromIncludeExclude(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, fromStdinFilter := ArgRootVariantFilterFromStdin(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		var compositeFilter = func(ctx *task.ExecutionContext, variant *dryad.SafeRootVariantReference) (error, bool) {
			var err error
			var shouldMatch bool

			err, shouldMatch = fromStdinFilter(ctx, variant)
			if err != nil {
				return err, false
			} else if !shouldMatch {
				return nil, false
			}

			err, shouldMatch = rootFilter(ctx, variant)
			return err, shouldMatch
		}

		return nil, ParsedArgs{
			Filter:            compositeFilter,
			VariantDescriptor: variantDescriptor,
			Path:              path,
			Parallel:          parallel,
			JoinStdout:        joinStdout,
			JoinStderr:        joinStderr,
			LogStdout:         logStdout,
			LogStderr:         logStderr,
		}
	}

	var buildGarden = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.Path)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = roots.Build(
			ctx,
			dryad.RootsBuildRequest{
				Filter:            args.Filter,
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

		return err, nil
	}

	buildGarden = task.WithContext(
		buildGarden,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			buildGarden,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while building garden")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("build", "build selected root variants in a garden").
		WithOption(
			clib.
				NewOption("path", "the target path for the garden to build").
				WithType(clib.OptionTypeString),
		).
		WithOption(
			clib.
				NewOption(
					"variant",
					"select a root variant descriptor for all matched roots (filesystem form: dimension=option+dimension=option). supports none/any/host; inherit is invalid. defaults to all enabled variants per root",
				).
				WithType(clib.OptionTypeString),
		).
		WithOption(clib.NewOption("include", "choose which root variants are included in the build. the include filter is a CEL expression with access to a 'root' object for each root variant.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which root variants are excluded from the build. the exclude filter is a CEL expression with access to a 'root' object for each root variant.").WithType(clib.OptionTypeMultiString)).
		WithOption(
			clib.NewOption(
				"from-stdin",
				"if set, read a list of root refs from stdin to use as a base list of root variants to build instead of all root variants. include and exclude filters will be applied after this list. default false",
			).
				WithType(clib.OptionTypeBool),
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

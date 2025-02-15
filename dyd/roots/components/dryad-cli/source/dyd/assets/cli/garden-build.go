package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	task "dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var gardenBuildCommand = func() clib.Command {

	type ParsedArgs struct {
		IncludeRoots func(path string) bool
		ExcludeRoots func(path string) bool
		Parallel int
		Path string
		JoinStdout bool
		JoinStderr bool
	}

	var parseArgs = func (ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		// var args = req.Args
		var options = req.Opts

		var path string
		// var err error

		var includeOpts []string
		var excludeOpts []string

		var parallel int

		var joinStdout bool
		var joinStderr bool

		if options["exclude"] != nil {
			excludeOpts = options["exclude"].([]string)
		}

		if options["include"] != nil {
			includeOpts = options["include"].([]string)
		}

		if options["path"] != nil {
			path = options["path"].(string)
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

		includeRoots := dryad.RootIncludeMatcher(includeOpts)
		excludeRoots := dryad.RootExcludeMatcher(excludeOpts)

		return nil, ParsedArgs{
			IncludeRoots: includeRoots,
			ExcludeRoots: excludeRoots,
			Path: path,
			Parallel: parallel,
			JoinStdout: joinStdout,
			JoinStderr: joinStderr,
		}
	}

	var buildGarden = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.Path)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, _ = dryad.RootsBuild(
			ctx,
			dryad.RootsBuildRequest{
				Garden:     garden,
				IncludeRoots: args.IncludeRoots,
				ExcludeRoots: args.ExcludeRoots,
				JoinStdout: args.JoinStdout,
				JoinStderr: args.JoinStderr,
			},
		)

		return err, nil
	}

	buildGarden = task.WithContext(
		buildGarden,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			buildGarden,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building garden")
				return 1
			}

			return 0
		},
	)
	
	command := clib.NewCommand("build", "build selected roots in a garden. alias for `dryad roots build`").
		WithOption(
			clib.
				NewOption("path", "the target path for the garden to build").
				WithType(clib.OptionTypeString),
		).
		WithOption(clib.NewOption("include", "choose which roots are included in the build").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the build").WithType(clib.OptionTypeMultiString)).
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
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

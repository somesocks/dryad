package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	tasks "dryad/tasks"

	zlog "github.com/rs/zerolog/log"
)

var gardenBuildCommand = func() clib.Command {

	type ParsedArgs struct {
		IncludeRoots func(path string) bool
		ExcludeRoots func(path string) bool
		Path string
	}

	var parseArgs = func(req clib.ActionRequest) (error, ParsedArgs) {
		// var args = req.Args
		var options = req.Opts

		var path string
		// var err error

		var includeOpts []string
		var excludeOpts []string

		if options["exclude"] != nil {
			excludeOpts = options["exclude"].([]string)
		}

		if options["include"] != nil {
			includeOpts = options["include"].([]string)
		}

		if options["path"] != nil {
			path = options["path"].(string)
		}


		includeRoots := dryad.RootIncludeMatcher(includeOpts)
		excludeRoots := dryad.RootExcludeMatcher(excludeOpts)

		return nil, ParsedArgs{
			IncludeRoots: includeRoots,
			ExcludeRoots: excludeRoots,
			Path: path,
		}
	}

	var buildGarden = func (args ParsedArgs) (error, any) {
		err := dryad.GardenBuild(
			dryad.BuildContext{
				Fingerprints: map[string]string{},
			},
			dryad.GardenBuildRequest{
				BasePath:     args.Path,
				IncludeRoots: args.IncludeRoots,
				ExcludeRoots: args.ExcludeRoots,
			},
		)

		return err, nil
	}

	var action = tasks.Return(
		tasks.Series2(
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
		WithAction(action)

	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

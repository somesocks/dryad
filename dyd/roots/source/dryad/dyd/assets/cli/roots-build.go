package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var rootsBuildCommand = func() clib.Command {
	command := clib.NewCommand("build", "build selected roots in a garden").
		WithOption(
			clib.
				NewOption("path", "the target path for the garden to build").
				WithType(clib.OptionTypeString),
		).
		WithOption(clib.NewOption("include", "choose which roots are included in the build").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the build").WithType(clib.OptionTypeMultiString)).
		WithAction(func(req clib.ActionRequest) int {
			// var args = req.Args
			var options = req.Opts

			var path string
			var err error

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

			err = dryad.GardenBuild(
				dryad.BuildContext{
					RootFingerprints: map[string]string{},
				},
				dryad.GardenBuildRequest{
					BasePath:     path,
					IncludeRoots: includeRoots,
					ExcludeRoots: excludeRoots,
				},
			)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		},
		)

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
)

var gardenBuildCommand = clib.NewCommand("build", "build all roots in the garden").
	WithArg(
		clib.
			NewArg("path", "the target path for the garden to build").
			AsOptional().
			WithAutoComplete(AutoCompletePath),
	).
	WithOption(clib.NewOption("include", "choose which roots are included in the build").WithType(clib.OptionTypeMultiString)).
	WithOption(clib.NewOption("exclude", "choose which roots are excluded from the build").WithType(clib.OptionTypeMultiString)).
	WithOption(clib.NewOption("scope", "set the scope for the command")).
	WithAction(scopeHandler(
		func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var path string
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			var includeOpts []string
			var excludeOpts []string

			if options["exclude"] != nil {
				excludeOpts = options["exclude"].([]string)
			}

			if options["include"] != nil {
				includeOpts = options["include"].([]string)
			}

			if len(args) > 0 {
				path = args[0]
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
	))

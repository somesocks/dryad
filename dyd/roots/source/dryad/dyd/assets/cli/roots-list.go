package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
)

var rootsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)").
		WithArg(
			clib.
				NewArg("path", "path to the base root (or garden) to list roots in").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("include", "choose which roots are included in the list").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the list").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("scope", "set the scope for the command")).
		WithAction(scopeHandler(
			func(req clib.ActionRequest) int {
				var args = req.Args
				var options = req.Opts

				var path string = ""
				var err error

				if len(args) > 0 {
					path = args[0]
				}

				var gardenPath string
				gardenPath, err = dryad.GardenPath(path)
				if err != nil {
					log.Fatal(err)
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				includeRoots := dryad.RootIncludeMatcher(includeOpts)
				excludeRoots := dryad.RootExcludeMatcher(excludeOpts)

				err = dryad.RootsWalk(path, func(path string, info fs.FileInfo) error {

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}

					if includeRoots(relPath) && !excludeRoots(relPath) {
						fmt.Println(path)
					}

					return nil
				})
				if err != nil {
					log.Fatal(err)
				}

				return 0
			},
		))

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

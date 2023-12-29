package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var sproutsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all sprouts of the current garden").
		WithOption(clib.NewOption("include", "choose which sprouts are included in the list").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded from the list").WithType(clib.OptionTypeMultiString)).
		WithAction(
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
					zlog.Fatal().Err(err)
					return 1
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				includeSprouts := dryad.RootIncludeMatcher(includeOpts)
				excludeSprouts := dryad.RootExcludeMatcher(excludeOpts)

				err = dryad.SproutsWalk(path, func(path string, info fs.FileInfo) error {

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}

					if includeSprouts(relPath) && !excludeSprouts(relPath) {
						fmt.Println(path)
					}

					return nil
				})
				if err != nil {
					zlog.Fatal().Err(err)
					return 1
				}

				return 0
			},
		)

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

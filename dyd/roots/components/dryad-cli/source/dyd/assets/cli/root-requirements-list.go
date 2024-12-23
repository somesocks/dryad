package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all requirements of this root").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts
			var root string

			if len(args) > 0 {
				root = args[0]
			}

			var relative bool = true

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			var gardenPath string
			gardenPath, err := dryad.GardenPath(root)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding garden path")
				return 1
			}

			err = dryad.RootRequirementsWalk(root, func(path string, info fs.FileInfo) error {
				if relative {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}
					fmt.Println(relPath)
				} else {
					fmt.Println(path)
				}
				return nil
			})

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling root requirements")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

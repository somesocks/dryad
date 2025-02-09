package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	fs2 "dryad/filesystem"
	"os"
	"path/filepath"
	// "regexp"

	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var stemAncestorsCommand = func() clib.Command {
	command := clib.NewCommand("ancestors", "list all direct and indirect dependencies of a stem").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print stems relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("self", "include the base stem itself. default false").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var err error
			var path string
			var relative bool = true
			var self bool = false

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while cleaning path")
					return 1
				}
			}
			if path == "" {
				path, err = os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
			}

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			if options["self"] != nil {
				self = options["self"].(bool)
			} else {
				self = false
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err = dryad.StemAncestorsWalk(
				dryad.StemAncestorsWalkRequest{
					BasePath: path,
					OnMatch: func(node fs2.Walk5Node) error {
						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(garden.BasePath, node.Path)
						if err != nil {
							return err
						}

						if relative {
							fmt.Println(relPath)
						} else {
							fmt.Println(node.Path)
						}

						return nil
					},
					Self: self,
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while listing files")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

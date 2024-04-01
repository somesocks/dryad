package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootsGraphCommand = func() clib.Command {
	command := clib.NewCommand("graph", "print the local dependency graph of all roots in the garden").
		WithOption(clib.NewOption("transpose", "transpose the dependency graph before printing").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var options = req.Opts

			var relative bool = true
			var transpose bool

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			if options["transpose"] != nil {
				transpose = options["transpose"].(bool)
			}

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			gardenPath, err := dryad.GardenPath(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding garden path")
				return 1
			}

			graph, err := dryad.RootsGraph(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building roots graph")
				return 1
			}

			if transpose {
				graph = graph.Transpose()
			}

			// Print the resulting roots
			if relative {
				for k, v := range graph {
					// calculate the relative path to the root from the base of the garden
					kPath, err := filepath.Rel(gardenPath, k)
					if err != nil {
						zlog.Fatal().Err(err).Msg("error while finding root")
						return 1
					}

					fmt.Println(kPath + ":")

					for _, vv := range v {
						// calculate the relative path to the root from the base of the garden
						vPath, err := filepath.Rel(gardenPath, vv)
						if err != nil {
							zlog.Fatal().Err(err).Msg("error while finding root")
							return 1
						}

						fmt.Println("  " + vPath)
					}

				}
			} else {
				for k, v := range graph {
					fmt.Println(k + ":")
					for _, vv := range v {
						fmt.Println("  " + vv)
					}
				}
			}

			return 0
		})

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"

	json "encoding/json"

	yaml "sigs.k8s.io/yaml"
)

var rootsGraphCommand = func() clib.Command {
	command := clib.NewCommand("graph", "print the local dependency graph of all roots in the garden").
		WithOption(clib.NewOption("transpose", "transpose the dependency graph before printing").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("format", "change the output format of the graph. can be one of (yaml, json, json-compact). defaults to yaml").WithType(clib.OptionTypeString)).
		WithAction(func(req clib.ActionRequest) int {
			var options = req.Opts

			var relative bool = true
			var format string = "yaml"
			var transpose bool

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			if options["transpose"] != nil {
				transpose = options["transpose"].(bool)
			}

			if options["format"] != nil {
				format = options["format"].(string)
				switch format {
				case "json", "JSON":
					format = "json"
					break
				case "json-compact", "JSON-COMPACT":
					format = "json-compact"
					break
				case "yaml", "YAML":
					format = "yaml"
					break
				default:
					zlog.
						Fatal().
						Str("format", format).
						Msg("unrecognized ouput format")
					return 1
				}
			}

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			graph, err := dryad.RootsGraph(path, relative)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building roots graph")
				return 1
			}

			if transpose {
				graph = graph.Transpose()
			}

			switch format {
			case "yaml":
				y, err := yaml.Marshal(graph)
				if err != nil {
					zlog.Fatal().Err(err).Msg("could not render graph to yaml")
					return 1
				}
				fmt.Println(string(y))
				break
			case "json":
				y, err := json.MarshalIndent(graph, "", "  ")
				if err != nil {
					zlog.Fatal().Err(err).Msg("could not render graph to json")
					return 1
				}
				fmt.Println(string(y))
				break
			case "json-compact":
				y, err := json.Marshal(graph)
				if err != nil {
					zlog.Fatal().Err(err).Msg("could not render graph to json-compact")
					return 1
				}
				fmt.Println(string(y))
				break
			}

			return 0
		})

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

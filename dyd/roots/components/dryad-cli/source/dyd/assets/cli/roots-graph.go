package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"

	json "encoding/json"

	yaml "sigs.k8s.io/yaml"
)

var rootsGraphCommand = func() clib.Command {
	type ParsedArgs struct {
		GardenPath string
		Relative   bool
		Format     string
		Transpose  bool
		Parallel   int
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var options = req.Opts

		var relative bool = true
		var format string = "yaml"
		var transpose bool
		var parallel int

		if options["relative"] != nil {
			relative = options["relative"].(bool)
		} else {
			relative = true
		}

		if options["transpose"] != nil {
			transpose = options["transpose"].(bool)
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		if options["format"] != nil {
			format = options["format"].(string)
			switch format {
			case "json", "JSON":
				format = "json"
			case "json-compact", "JSON-COMPACT":
				format = "json-compact"
			case "yaml", "YAML":
				format = "yaml"
			default:
				return fmt.Errorf("unrecognized output format: %s", format), ParsedArgs{}
			}
		}

		path, err := os.Getwd()
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			GardenPath: path,
			Relative:   relative,
			Format:     format,
			Transpose:  transpose,
			Parallel:   parallel,
		}
	}

	var printGraph = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, graph := roots.Graph(
			ctx,
			dryad.RootsGraphRequest{
				Relative: args.Relative,
			},
		)
		if err != nil {
			return err, nil
		}

		if args.Transpose {
			graph = graph.Transpose()
		}

		switch args.Format {
		case "yaml":
			y, err := yaml.Marshal(graph)
			if err != nil {
				return err, nil
			}
			fmt.Println(string(y))
		case "json":
			y, err := json.MarshalIndent(graph, "", "  ")
			if err != nil {
				return err, nil
			}
			fmt.Println(string(y))
		case "json-compact":
			y, err := json.Marshal(graph)
			if err != nil {
				return err, nil
			}
			fmt.Println(string(y))
		}

		return nil, nil
	}

	printGraph = task.WithContext(
		printGraph,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printGraph,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building roots graph")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("graph", "print the local dependency graph of all roots in the garden").
		WithOption(clib.NewOption("transpose", "transpose the dependency graph before printing").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("format", "change the output format of the graph. can be one of (yaml, json, json-compact). defaults to yaml").WithType(clib.OptionTypeString)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

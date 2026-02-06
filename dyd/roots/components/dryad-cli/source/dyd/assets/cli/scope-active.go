package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeActiveCommand = func() clib.Command {
	type ParsedArgs struct {
		Oneline    bool
		GardenPath string
		Parallel   int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var opts = req.Opts
			var oneline bool = true
			var parallel int

			if opts["oneline"] != nil {
				oneline = opts["oneline"].(bool)
			}

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			var path, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				Oneline:    oneline,
				GardenPath: path,
				Parallel:   parallel,
			}
		},
	)

	var printActiveScope = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		scopeName, err := dryad.ScopeGetDefault(garden)
		if err != nil {
			return err, nil
		}

		if scopeName == "" {
			return nil, nil
		}

		var scopeOneline string = ""
		if args.Oneline {
			scopeOneline, _ = dryad.ScopeOnelineGet(garden, scopeName)
		}

		if scopeOneline != "" {
			scopeName = scopeName + " - " + scopeOneline
		}

		fmt.Println(scopeName)

		return nil, nil
	}

	printActiveScope = task.WithContext(
		printActiveScope,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printActiveScope,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while loading active scope")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("active", "return the name of the active scope, if set. alias for `dryad scopes default get`").
		WithOption(clib.NewOption("oneline", "enable/disable printing one-line scope descriptions").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

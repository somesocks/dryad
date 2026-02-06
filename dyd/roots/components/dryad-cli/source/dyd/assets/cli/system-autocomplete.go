package cli

import (
	clib "dryad/cli-builder"
	"dryad/task"
	"fmt"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var systemAutocomplete = func() clib.Command {
	type ParsedArgs struct {
		Parallel  int
		Args      []string
		Separator string
		App       clib.App
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var opts = req.Opts
		var separator string
		var parallel int

		if opts["separator"] != nil {
			separator = opts["separator"].(string)
		} else {
			separator = " "
		}

		if opts["parallel"] != nil {
			parallel = int(opts["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Parallel:  parallel,
			Args:      req.Args[0:],
			Separator: separator,
			App:       req.App,
		}
	}

	var runAutocomplete = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		var err, results = args.App.AutoComplete(args.Args)
		if err != nil {
			return err, nil
		}

		fmt.Println(strings.Join(results, args.Separator))

		return nil, nil
	}

	runAutocomplete = task.WithContext(
		runAutocomplete,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runAutocomplete,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building autocomplete tokens")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("autocomplete", "print out autocomplete options based on a partial command").
		WithArg(clib.NewArg("-- args", "args to pass to the command").AsOptional()).
		WithOption(clib.NewOption("separator", "separator string to use between tokens.")).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

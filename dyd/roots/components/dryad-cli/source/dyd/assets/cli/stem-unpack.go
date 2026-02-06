package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemUnpackCommand = func() clib.Command {
	type ParsedArgs struct {
		GardenPath string
		StemPath   string
		Parallel   int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts
		var parallel int

		gardenPath, err := os.Getwd()
		if err != nil {
			return err, ParsedArgs{}
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			GardenPath: gardenPath,
			StemPath:   args[0],
			Parallel:   parallel,
		}
	}

	var runUnpack = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		targetPath, err := dryad.StemUnpack(args.GardenPath, args.StemPath)
		if err != nil {
			return err, nil
		}

		fmt.Println(targetPath)
		return nil, nil
	}

	runUnpack = task.WithContext(
		runUnpack,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runUnpack,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while unpacking stem")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("unpack", "unpack a stem archive at the target path and import it into the current garden").
		WithArg(
			clib.
				NewArg("archive", "the path to the archive to unpack").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

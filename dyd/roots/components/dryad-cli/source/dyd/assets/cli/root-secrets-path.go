package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootSecretsPathCommand = func() clib.Command {
	type ParsedArgs struct {
		Path     string
		Parallel int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts
			var err error
			var path string

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					return err, ParsedArgs{}
				}
			} else {
				path, err = os.Getwd()
				if err != nil {
					return err, ParsedArgs{}
				}
			}

			var parallel int
			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			return nil, ParsedArgs{
				Path:     path,
				Parallel: parallel,
			}
		},
	)

	var printSecretsPath = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		path := args.Path

		// normalize the path to point to the closest secrets
		path, err := dryad.SecretsPath(path)
		if err != nil {
			return err, nil
		}

		// check if the secrets folder exists
		exists, err := dryad.SecretsExist(path)
		if err != nil {
			return err, nil
		}

		if exists {
			fmt.Println(path)
		}

		return nil, nil
	}

	printSecretsPath = task.WithContext(
		printSecretsPath,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			printSecretsPath,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding secrets path")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("path", "print the path to the secrets for the current package, if it exists").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

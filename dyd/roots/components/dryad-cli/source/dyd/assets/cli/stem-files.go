package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

var stemFilesCommand = func() clib.Command {
	type ParsedArgs struct {
		Path         string
		MatchExclude *regexp.Regexp
		Parallel     int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var err error
		var path string
		var matchExclude *regexp.Regexp
		var parallel int

		if options["exclude"] != nil {
			exclude := options["exclude"].(string)
			if exclude != "" {
				matchExclude, err = regexp.Compile(exclude)
				if err != nil {
					return err, ParsedArgs{}
				}
			}
		}

		if len(args) > 0 {
			path = args[0]
			path, err = filepath.Abs(path)
			if err != nil {
				return err, ParsedArgs{}
			}
		}

		if path == "" {
			path, err = os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Path:         path,
			MatchExclude: matchExclude,
			Parallel:     parallel,
		}
	}

	var runStemFiles = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err := dryad.StemFiles(
			ctx,
			dryad.StemFilesArgs{
				BasePath:  args.Path,
				MatchDeny: args.MatchExclude,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	runStemFiles = task.WithContext(
		runStemFiles,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runStemFiles,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while listing files")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("files", "list the files in a stem").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("exclude", "a regular expression to exclude files from the list. the regexp matches against the file path relative to the stem base directory")).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

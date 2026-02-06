package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

var stemFingerprintCommand = func() clib.Command {
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

	var runFingerprint = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, fingerprintString := dryad.StemFingerprint(
			ctx,
			dryad.StemFingerprintRequest{
				BasePath:  args.Path,
				MatchDeny: args.MatchExclude,
			},
		)
		if err != nil {
			return err, nil
		}

		fmt.Println(fingerprintString)
		return nil, nil
	}

	runFingerprint = task.WithContext(
		runFingerprint,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runFingerprint,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building stem fingerprint")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("exclude", "a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory")).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

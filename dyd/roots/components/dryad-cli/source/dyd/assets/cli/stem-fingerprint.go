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
	command := clib.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("exclude", "a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var err error
			var matchExclude *regexp.Regexp

			if options["exclude"] != "" {
				matchExclude, err = regexp.Compile(options["exclude"].(string))
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while compiling exclusion expression")
					return 1
				}
			}

			var path string
			if len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				path, err = os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
			}

			var fingerprintErr, fingerprintString = dryad.StemFingerprint(
				task.SERIAL_CONTEXT,
				dryad.StemFingerprintRequest{
					BasePath:  path,
					MatchDeny: matchExclude,
				},
			)
			if fingerprintErr != nil {
				zlog.Fatal().Err(fingerprintErr).Msg("error while building stem fingerprint")
				return 1
			}
			fmt.Println(fingerprintString)

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

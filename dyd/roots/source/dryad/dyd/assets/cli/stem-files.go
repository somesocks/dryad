package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

var stemFilesCommand = func() clib.Command {
	command := clib.NewCommand("files", "list the files in a stem").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("exclude", "a regular expression to exclude files from the list. the regexp matches against the file path relative to the stem base directory")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var err error
			var matchExclude *regexp.Regexp

			if options["exclude"] != nil && options["exclude"] != "" {
				matchExclude, err = regexp.Compile(options["exclude"].(string))
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while compiling exclusion expression")
					return -1
				}
			}

			var path string
			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while cleaning path")
					return 1
				}
			}
			if path == "" {
				path, err = os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
			}

			err = dryad.StemFiles(
				dryad.StemFilesArgs{
					BasePath:  path,
					MatchDeny: matchExclude,
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while listing files")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

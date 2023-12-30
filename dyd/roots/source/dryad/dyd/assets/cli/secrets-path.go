package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var secretsPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "print the path to the secrets for the current package, if it exists").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var err error
			var path string

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while cleaning path")
					return 1
				}
			} else {
				path, err = os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
			}

			// normalize the path to point to the closest secrets
			path, err = dryad.SecretsPath(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding secrets path")
				return 1
			}

			// check if the secrets folder exists
			exists, err := dryad.SecretsExist(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error checking if secrets exist")
				return 1
			}

			if exists {
				fmt.Println(path)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

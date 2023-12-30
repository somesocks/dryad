package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var secretsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list the secret files in a stem/root").
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

			err = dryad.SecretsWalk(
				dryad.SecretsWalkArgs{
					BasePath: path,
					OnMatch: func(path string, info fs.FileInfo) error {
						fmt.Println(path)
						return nil
					},
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling secrets")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootBuildCommand = func() clib.Command {
	command := clib.
		NewCommand("build", "build a specified root").
		WithArg(
			clib.
				NewArg("path", "path to the root to build").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string

			if len(args) > 0 {
				path = args[0]
			}

			if !filepath.IsAbs(path) {
				wd, err := os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
				path = filepath.Join(wd, path)
			}

			var rootFingerprint string
			rootFingerprint, err := dryad.RootBuild(
				dryad.BuildContext{
					Fingerprints: map[string]string{},
				},
				path,
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building root")
				return 1
			}
			fmt.Println(rootFingerprint)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

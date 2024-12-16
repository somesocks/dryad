package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the base path of the current root").
		WithArg(
			clib.
				NewArg("path", "the path to start searching for a root at. defaults to current directory").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var path string = ""

			if len(args) > 0 {
				path = args[0]
			}

			path, err := dryad.RootPath(path, "")
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding root path")
				return 1
			}
			fmt.Println(path)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

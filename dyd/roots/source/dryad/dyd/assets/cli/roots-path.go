package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootsPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the path of the roots dir").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}
			path, err = dryad.RootsPath(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding roots path")
				return 1
			}
			fmt.Println(path)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

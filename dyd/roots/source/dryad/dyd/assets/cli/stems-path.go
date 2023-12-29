package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemsPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the path of the stems dir").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}
			path, err = dryad.StemsPath(path)
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}
			fmt.Println(path)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

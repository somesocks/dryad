package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var stemPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "return the base path of the current root").
		// WithArg(clib.NewArg("path", "path to the stem base dir")).
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err)
				return 1
			}
			path, err = dryad.StemPath(path)
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

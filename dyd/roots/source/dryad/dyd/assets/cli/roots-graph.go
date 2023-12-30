package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootsGraphCommand = func() clib.Command {
	command := clib.NewCommand("graph", "print the local dependency graph of all roots in the garden").
		WithAction(func(req clib.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			graph, err := dryad.RootsGraph(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building roots graph")
				return 1
			}

			for k, v := range graph {
				fmt.Println(k + ":")
				for _, vv := range v {
					fmt.Println("  " + vv)
				}
			}

			return 0
		})

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

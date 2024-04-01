package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootDescendantsCommand = func() clib.Command {
	command := clib.NewCommand("descendants", "list all roots that depend on the selected root (directly and indirectly)").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var rootPath string

			if len(args) > 0 {
				rootPath = args[0]
			}

			if !filepath.IsAbs(rootPath) {
				wd, err := os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
				rootPath = filepath.Join(wd, rootPath)
			}

			gardenPath, err := dryad.GardenPath(rootPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding garden path")
				return 1
			}

			graph, err := dryad.RootsGraph(gardenPath, false)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building graph")
				return 1
			}

			graph = graph.Transpose()

			ancestors := graph.Descendants(make(dryad.TStringSet), []string{rootPath}).ToArray([]string{})
			for _, v := range ancestors {
				fmt.Println(v)
			}

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

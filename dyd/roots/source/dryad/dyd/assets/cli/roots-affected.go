package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootsAffectedCommand = func() clib.Command {
	command := clib.NewCommand("affected", "take a list of files from stdin, and print a list of roots that may depend on those files").
		WithAction(func(req clib.ActionRequest) int {

			wd, err := os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			rootSet := make(dryad.TStringSet)

			scanner := bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				path := scanner.Text()
				path, err := filepath.Abs(path)
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while reading from stdin")
					return 1
				}
				path = _rootsOwningDependencyCorrection(path)
				path, err = dryad.RootPath(path)
				if err == nil {
					rootSet[path] = true
				}
			}

			// Check for any errors during scanning
			if err := scanner.Err(); err != nil {
				zlog.Fatal().Err(err).Msg("error after reading from stdin")
				return 1
			}

			rootList := rootSet.ToArray([]string{})

			gardenPath, err := dryad.GardenPath(wd)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding garden path")
				return 1
			}

			graph, err := dryad.RootsGraph(gardenPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building roots graph")
				return 1
			}

			graph = graph.Transpose()

			// find the descendants of the affected roots
			descendants := graph.Descendants(make(dryad.TStringSet), rootList)
			for k := range descendants {
				rootSet[k] = true
			}

			// print all of the resulting roots
			for key := range rootSet {
				fmt.Println(key)
			}

			return 0
		})

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

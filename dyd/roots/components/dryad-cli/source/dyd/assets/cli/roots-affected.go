package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootsAffectedCommand = func() clib.Command {
	command := clib.NewCommand("affected", "take a list of files from stdin, and print a list of roots that may depend on those files").
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var options = req.Opts

			var relative bool = true

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

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
				path, err = dryad.RootPath(path, "")
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

			unsafeGarden := dryad.Garden(wd)
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden roots")
				return 1
			}

			err, graph := roots.Graph(
				task.SERIAL_CONTEXT,
				dryad.RootsGraphRequest{
					Relative: false,
				},
			)
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

			// Print the resulting roots
			if relative {
				for key := range rootSet {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(garden.BasePath, key)
					if err != nil {
						zlog.Fatal().Err(err).Msg("error while finding root")
						return 1
					}
					fmt.Println(relPath)
				}
			} else {
				for key := range rootSet {
					fmt.Println(key)
				}
			}

			return 0
		})

	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()

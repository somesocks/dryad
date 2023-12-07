package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var rootsAffectedCommand = func() clib.Command {
	command := clib.NewCommand("affected", "take a list of files from stdin, and print a list of roots that may depend on those files").
		WithAction(func(req clib.ActionRequest) int {

			wd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			rootSet := make(dryad.TStringSet)

			scanner := bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				path := scanner.Text()
				path, err := filepath.Abs(path)
				if err != nil {
					log.Fatal(err)
				}
				path = _rootsOwningDependencyCorrection(path)
				path, err = dryad.RootPath(path)
				if err == nil {
					rootSet[path] = true
				}
			}

			// Check for any errors during scanning
			if err := scanner.Err(); err != nil {
				log.Fatal("error reading stdin", err)
			}

			rootList := rootSet.ToArray([]string{})

			gardenPath, err := dryad.GardenPath(wd)
			if err != nil {
				log.Fatal(err)
			}

			graph, err := dryad.RootsGraph(gardenPath)
			if err != nil {
				log.Fatal(err)
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

package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var rootsOwningCommand = func() clib.Command {
	command := clib.NewCommand("owning", "list all roots that are owners of the provided files. The files to check should be provided as relative or absolute paths through stdin.").
		WithAction(
			func(req clib.ActionRequest) int {

				rootSet := make(map[string]bool)

				scanner := bufio.NewScanner(os.Stdin)

				for scanner.Scan() {
					filePath := scanner.Text()
					rootPath, err := dryad.RootPath(filePath)
					if err == nil {
						rootSet[rootPath] = true
					}
				}

				// Check for any errors during scanning
				if err := scanner.Err(); err != nil {
					log.Fatal("error reading stdin", err)
				}

				// Print the resulting roots
				for key := range rootSet {
					fmt.Println(key)
				}

				return 0
			},
		)

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

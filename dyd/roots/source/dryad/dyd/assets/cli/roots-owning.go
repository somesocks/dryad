package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/rs/zerolog/log"
)

var _rootsOwningDependencyCorrection = func(path string) string {
	p1, _ := filepath.Split(path)
	p1 = filepath.Clean(p1)
	p2, f2 := filepath.Split(p1)
	p2 = filepath.Clean(p2)
	p3, f3 := filepath.Split(p2)
	p3 = filepath.Clean(p3)

	if f3 == "dyd" && f2 == "requirements" {
		return p3
	} else {
		return path
	}

}

var rootsOwningCommand = func() clib.Command {
	command := clib.NewCommand("owning", "list all roots that are owners of the provided files. The files to check should be provided as relative or absolute paths through stdin.").
		WithAction(
			func(req clib.ActionRequest) int {

				rootSet := make(map[string]bool)

				scanner := bufio.NewScanner(os.Stdin)

				for scanner.Scan() {
					path := scanner.Text()
					path, err := filepath.Abs(path)
					if err != nil {
						log.Error().
							Err(err).
							Msg("")
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
					// log.Fatal("error reading stdin", err)
					return 1
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

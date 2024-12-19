package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	zlog "github.com/rs/zerolog/log"
)

var scopesListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all scopes in the current garden").
		WithOption(clib.NewOption("oneline", "enable/disable printing one-line scope descriptions").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var opts = req.Opts

			var path string = ""
			var err error
			var oneline bool = true

			if opts["oneline"] != nil {
				oneline = opts["oneline"].(bool)
			}

			if len(args) > 0 {
				path = args[0]
			}

			var scopes []string

			err = dryad.ScopesWalk(path, func(path string, info fs.FileInfo) error {
				scope := filepath.Base(path)

				// fetch the oneline if enabled
				if oneline {
					scopeOneline, _ := dryad.ScopeOnelineGet(path, scope)
					if scopeOneline != "" {
						scope = scope + " - " + scopeOneline
					}
				}

				scopes = append(scopes, scope)

				return nil
			})
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling scopes")
				return 1
			}

			sort.Strings(scopes)

			for _, scope := range scopes {
				fmt.Println(scope)
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scopeActiveCommand = func() clib.Command {
	command := clib.NewCommand("active", "return the name of the active scope, if set. alias for `dryad scopes default get`").
		WithOption(clib.NewOption("oneline", "enable/disable printing one-line scope descriptions").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var opts = req.Opts

			var oneline bool = true

			if opts["oneline"] != nil {
				oneline = opts["oneline"].(bool)
			}

			var path, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				return 1
			}
		
			scopeName, err := dryad.ScopeGetDefault(&garden)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while loading active scope")
				return 1
			}

			if scopeName == "" {
				return 0
			}

			var scopeOneline string = ""
			if oneline {
				scopeOneline, _ = dryad.ScopeOnelineGet(&garden, scopeName)
			}

			if scopeOneline != "" {
				scopeName = scopeName + " - " + scopeOneline
			}

			fmt.Println(scopeName)

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

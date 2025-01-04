package cli

import (
	clib "dryad/cli-builder"
	// dryad "dryad/core"
	// "os"
	// "strings"

	zlog "github.com/rs/zerolog/log"
)

var ParallelCommand = func(
	command clib.Command,
) clib.Command {
	action := command.Action()

	wrapper := func(req clib.ActionRequest) int {
		// invocation := req.Invocation
		options := req.Opts

		var parallel int

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
			if parallel < 1 || parallel > 64 {
				zlog.
					Fatal().
					Msg("invalid number of parallel tasks specified, must be between 1 and 64")
				return 1
			}
		}


		return action(req)
	}

	return command.
		WithOption(
			clib.
			NewOption("parallel", "the number of actions to run in parallel").
			WithType(clib.OptionTypeInt),
		).
		WithAction(wrapper)
}

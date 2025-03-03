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
			if parallel < PARALLEL_COUNT_MIN || parallel > PARALLEL_COUNT_MAX {
				zlog.
					Fatal().
					Msg("invalid number of parallel threads specified, must be between 1 and 64")
				return 1
			}
		}


		return action(req)
	}

	return command.
		WithOption(
			clib.
			NewOption("parallel", "set the number of threads used to execute tasks in parallel. set to 1 to execute all tasks serially").
			WithType(clib.OptionTypeInt),
		).
		WithAction(wrapper)
}

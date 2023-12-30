package cli

import (
	clib "dryad/cli-builder"
	"fmt"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var systemAutocomplete = func() clib.Command {
	command := clib.NewCommand("autocomplete", "print out autocomplete options based on a partial command").
		WithArg(clib.NewArg("-- args", "args to pass to the command").AsOptional()).
		WithOption(clib.NewOption("separator", "separator string to use between tokens.")).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args[0:]
			var options = req.Opts

			var separator string

			if options["separator"] != nil {
				separator = options["separator"].(string)
			} else {
				separator = " "
			}

			var err, results = req.App.AutoComplete(args)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building autocomplete tokens")
				return 1
			}
			fmt.Println(strings.Join(results, separator))
			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

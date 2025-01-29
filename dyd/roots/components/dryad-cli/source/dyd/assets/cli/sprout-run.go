package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var sproutRunCommand = func() clib.Command {
	command := clib.NewCommand("run", "execute the main for a sprout").
		WithArg(
			clib.
				NewArg("path", "path to the sprout").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("override", "run this executable in the stem run envinronment instead of the main")).
		WithOption(clib.NewOption("confirm", "ask for a confirmation string to be entered to execute this command").WithType(clib.OptionTypeString)).
		WithArg(clib.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var context string
			var override string
			var inherit bool
			var confirm string

			if options["context"] != nil {
				context = options["context"].(string)
			}

			if options["inherit"] != nil {
				inherit = options["inherit"].(bool)
			}

			if options["override"] != nil {
				override = options["override"].(string)
			}

			if options["confirm"] != nil {
				confirm = options["confirm"].(string)
			}

			// if confirm is set, we want to print the list
			// of sprouts to run
			if confirm != "" {
				fmt.Println("this package will be executed:")
				fmt.Println(args[0])
				fmt.Println("are you sure? type '" + confirm + "' to continue")

				reader := bufio.NewReader(os.Stdin)

				input, err := reader.ReadString('\n')
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while reading input")
					return 1
				}

				input = strings.TrimSuffix(input, "\n")

				if input != confirm {
					zlog.Fatal().Msg("input does not match confirmation, aborting")
					return 1
				}

			}

			var env = map[string]string{}

			// pull
			if inherit {
				for _, e := range os.Environ() {
					if i := strings.Index(e, "="); i >= 0 {
						env[e[:i]] = e[i+1:]
					}
				}
			} else {
				// copy a few variables over from parent env for convenience
				env["TERM"] = os.Getenv("TERM")
			}

			path := args[0]
			extras := args[1:]
			err := dryad.StemRun(dryad.StemRunRequest{
				StemPath:     path,
				MainOverride: override,
				Env:          env,
				Args:         extras,
				JoinStdout:   true,
				JoinStderr:   true,
				Context:      context,
			})
			if err != nil {
				zlog.Fatal().Err(err).Msg("error executing stem")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

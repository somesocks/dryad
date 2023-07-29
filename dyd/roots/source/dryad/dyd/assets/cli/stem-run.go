package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
	"strings"
)

var stemRunCommand = clib.NewCommand("run", "execute the main for a stem").
	WithArg(clib.NewArg("path", "path to the stem base dir")).
	WithOption(clib.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
	WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(clib.OptionTypeBool)).
	WithOption(clib.NewOption("override", "run this executable in the stem run envinronment instead of the main")).
	WithArg(clib.NewArg("-- args", "args to pass to the stem").AsOptional()).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args
		var options = req.Opts

		var context string
		var override string
		var inherit bool

		if options["context"] != nil {
			context = options["context"].(string)
		}

		if options["inherit"] != nil {
			inherit = options["inherit"].(bool)
		}

		if options["override"] != nil {
			override = options["override"].(string)
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
			Context:      context,
		})
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})

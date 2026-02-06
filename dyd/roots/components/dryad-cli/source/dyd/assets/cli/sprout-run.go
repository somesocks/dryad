package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func resolveSproutPath(path string) (error, string) {
	// Resolve symlinks in the parent path, but preserve the final segment.
	// This keeps sprout symlinks under dyd/sprouts intact while still
	// normalizing paths like /tmp -> /private/tmp on macOS.
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return err, ""
	}

	parent := filepath.Dir(absPath)
	base := filepath.Base(absPath)

	parent, err = filepath.EvalSymlinks(parent)
	if err != nil {
		return err, ""
	}

	return nil, filepath.Join(parent, base)
}

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
		WithOption(clib.NewOption("join-stdout", "join the stdout of child processes to the stderr of the parent dryad process. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("join-stderr", "join the stderr of child processes to the stderr of the parent dryad process. default true").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("log-stdout", "log the stdout of child processes to a file in the specified directory. disables joining").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("log-stderr", "log the stderr of child processes to a file in the specified directory. disables joining").WithType(clib.OptionTypeString)).
		WithArg(clib.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var context string
			var override string
			var inherit bool
			var confirm string
			var joinStdout bool
			var joinStderr bool
			var logStdout string
			var logStderr string

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

			if options["join-stdout"] != nil {
				joinStdout = options["join-stdout"].(bool)
			} else {
				joinStdout = true
			}

			if options["join-stderr"] != nil {
				joinStderr = options["join-stderr"].(bool)
			} else {
				joinStderr = true
			}

			if options["log-stdout"] != nil {
				logStdout = options["log-stdout"].(string)
				joinStdout = false
			}

			if options["log-stderr"] != nil {
				logStderr = options["log-stderr"].(string)
				joinStderr = false
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

			var err error
			path := args[0]
			extras := args[1:]

			err, path = resolveSproutPath(path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving path")
				return 1
			}
			
			err, garden := dryad.Garden(path).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err, sprouts := garden.Sprouts().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving sprouts")
				return 1
			}

			err, sprout := sprouts.Sprout(path).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving sprout")
				return 1
			}

			err = sprout.Run(
				task.SERIAL_CONTEXT,
				dryad.SproutRunRequest{
					MainOverride: override,
					Env:          env,
					Args:         extras,
					JoinStdout:   joinStdout,
					LogStdout:    struct {
						Path string
						Name string
					}{
						Path: logStdout,
						Name: "",
					},
					JoinStderr:   joinStderr,
					LogStderr:    struct {
						Path string
						Name string
					}{
						Path: logStderr,
						Name: "",
					},
					Context:      context,
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error executing stem")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()

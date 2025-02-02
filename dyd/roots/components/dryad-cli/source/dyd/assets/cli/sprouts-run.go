package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var sproutsRunCommand = func() clib.Command {
	command := clib.NewCommand("run", "run each sprout in the current garden").
		WithOption(clib.NewOption("include", "choose which sprouts are included").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("confirm", "ask for a confirmation string to be entered to execute this command").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("ignore-errors", "continue running even if a sprout returns an error").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("join-stdout", "join the stdout of child processes to the stderr of the parent dryad process. default false").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("join-stderr", "join the stderr of child processes to the stderr of the parent dryad process. default false").WithType(clib.OptionTypeBool)).
		WithArg(clib.NewArg("-- args", "args to pass to each sprout on execution").AsOptional()).
		WithAction(
			func(req clib.ActionRequest) int {
				var args = req.Args
				var options = req.Opts

				var path string = ""
				var err error

				if len(args) > 0 {
					path = args[0]
				}

				var gardenPath string
				gardenPath, err = dryad.GardenPath(path)
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding garden path")
					return 1
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				includeSprouts := dryad.RootIncludeMatcher(includeOpts)
				excludeSprouts := dryad.RootExcludeMatcher(excludeOpts)

				var context string
				var inherit bool
				var ignoreErrors bool
				var confirm string
				var joinStdout bool
				var joinStderr bool
				
				if options["context"] != nil {
					context = options["context"].(string)
				}

				if options["inherit"] != nil {
					inherit = options["inherit"].(bool)
				}

				if options["ignore-errors"] != nil {
					ignoreErrors = options["ignore-errors"].(bool)
				}

				if options["join-stdout"] != nil {
					joinStdout = options["join-stdout"].(bool)
				} else {
					joinStdout = false
				}
		
				if options["join-stderr"] != nil {
					joinStderr = options["join-stderr"].(bool)
				} else {
					joinStderr = false
				}		

				if options["confirm"] != nil {
					confirm = options["confirm"].(string)
				}

				// if confirm is set, we want to print the list
				// of sprouts to run
				if confirm != "" {
					fmt.Println("dryad sprouts exec will execute these sprouts:")

					err = dryad.SproutsWalk(path, func(path string, info fs.FileInfo) error {

						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(gardenPath, path)
						if err != nil {
							return err
						}

						if includeSprouts(relPath) && !excludeSprouts(relPath) {
							fmt.Println(" - " + path)
						}

						return nil
					})
					if err != nil {
						zlog.Fatal().Err(err).Msg("error while crawling sprouts")
						return 1
					}

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

				// pull environment variables from parent process
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

				extras := args[0:]

				err = dryad.SproutsWalk(path, func(path string, info fs.FileInfo) error {

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}

					if includeSprouts(relPath) && !excludeSprouts(relPath) {
						zlog.Info().
							Str("sprout", path).
							Msg("sprout run starting")

						err := dryad.StemRun(dryad.StemRunRequest{
							StemPath:   path,
							Env:        env,
							Args:       extras,
							JoinStdout: joinStdout,
							JoinStderr: joinStderr,
							Context:    context,
						})
						if err != nil {
							zlog.Warn().
								Str("sprout", path).
								Err(err).
								Msg("sprout threw error during execution")
							if !ignoreErrors {
								return err
							}
						} else {
							zlog.Info().
								Str("sprout", path).
								Msg("sprout run finished")
						}

					}

					return nil
				})
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while crawling sprouts")
					return 1
				}

				return 0
			},
		)

	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()

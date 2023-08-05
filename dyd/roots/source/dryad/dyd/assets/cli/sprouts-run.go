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

	log "github.com/rs/zerolog/log"
)

var sproutsRunCommand = func() clib.Command {
	command := clib.NewCommand("run", "run each sprout in the current garden").
		WithOption(clib.NewOption("include", "choose which sprouts are included").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("confirm", "display the list of sprouts to exec, and ask for confirmation").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("ignore-errors", "continue running even if a sprout returns an error").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("scope", "set the scope for the command")).
		WithArg(clib.NewArg("-- args", "args to pass to each sprout on execution").AsOptional()).
		WithAction(scopeHandler(
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
					log.Fatal().Err(err)
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
				var confirm bool

				if options["context"] != nil {
					context = options["context"].(string)
				}

				if options["inherit"] != nil {
					inherit = options["inherit"].(bool)
				}

				if options["ignore-errors"] != nil {
					ignoreErrors = options["ignore-errors"].(bool)
				}

				if options["confirm"] != nil {
					confirm = options["confirm"].(bool)
				}

				// if confirm is set, we want to print the list
				// of sprouts to run
				if confirm {
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
						log.Fatal().Err(err)
					}

					fmt.Println("are you sure? y/n")

					reader := bufio.NewReader(os.Stdin)

					input, err := reader.ReadString('\n')
					if err != nil {
						log.Fatal().Err(err).Msg("error reading input")
						return -1
					}

					input = strings.TrimSuffix(input, "\n")

					if input != "y" {
						fmt.Println("confirmation denied, aborting")
						return 0
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
						log.Info().Msg("running sprout at " + path)

						err := dryad.StemRun(dryad.StemRunRequest{
							StemPath:   path,
							Env:        env,
							Args:       extras,
							JoinStdout: true,
							Context:    context,
						})
						if err != nil {
							if ignoreErrors {
								log.Warn().Msg("sprout at " + path + " threw error ")
							} else {
								return err
							}
						}

					}

					return nil
				})
				if err != nil {
					log.Fatal().Err(err)
				}

				return 0
			},
		))

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

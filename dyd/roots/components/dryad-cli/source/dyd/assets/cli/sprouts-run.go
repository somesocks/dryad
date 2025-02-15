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

var sproutsRunCommand = func() clib.Command {

	type ParsedArgs struct {
		GardenPath string
		Parallel int
		IncludeSprouts func(path string) bool
		ExcludeSprouts func(path string) bool
		Confirm string
		Context string
		Inherit bool
		IgnoreErrors bool
		JoinStdout bool
		JoinStderr bool
		Extras []string
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

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
			var parallel int
			
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

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}
	
			if options["confirm"] != nil {
				confirm = options["confirm"].(string)
			}

			extras := args[0:]

			return nil, ParsedArgs{
				GardenPath: "",
				Parallel: parallel,
				IncludeSprouts: includeSprouts,
				ExcludeSprouts: excludeSprouts,
				Confirm: confirm,
				Context: context,
				Inherit: inherit,
				IgnoreErrors: ignoreErrors,
				JoinStdout: joinStdout,
				JoinStderr: joinStderr,
				Extras: extras,
			}
		},
	)

	var runSprouts = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {

		unsafeGarden := dryad.Garden(args.GardenPath)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		// if confirm is set, we want to print the list
		// of sprouts to run
		if args.Confirm != "" {
			fmt.Println("dryad sprouts exec will execute these sprouts:")

			err, _ := dryad.SproutsWalk(
				ctx,
				dryad.SproutsWalkRequest{
					Garden: garden,
					OnSprout: func (ctx *task.ExecutionContext, path string) (error, any) {
						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(garden.BasePath, path)
						if err != nil {
							return err, nil
						}

						if args.IncludeSprouts(relPath) && !args.ExcludeSprouts(relPath) {
							fmt.Println(" - " + path)
						}

						return nil, nil
					},
				},
			)
				
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling sprouts")
				return err, nil
			}

			fmt.Println("are you sure? type '" + args.Confirm + "' to continue")

			reader := bufio.NewReader(os.Stdin)

			input, err := reader.ReadString('\n')
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while reading input")
				return err, nil
			}

			input = strings.TrimSuffix(input, "\n")

			if input != args.Confirm {
				zlog.Fatal().Msg("input does not match confirmation, aborting")
				return err, nil
			}

		}


		var env = map[string]string{}

		// pull environment variables from parent process
		if args.Inherit {
			for _, e := range os.Environ() {
				if i := strings.Index(e, "="); i >= 0 {
					env[e[:i]] = e[i+1:]
				}
			}
		} else {
			// copy a few variables over from parent env for convenience
			env["TERM"] = os.Getenv("TERM")
		}

		err, _ = dryad.SproutsWalk(
			ctx,
			dryad.SproutsWalkRequest{
				Garden: garden,
				OnSprout: func (ctx *task.ExecutionContext, path string) (error, any) {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(garden.BasePath, path)
					if err != nil {
						return err, nil
					}

					if args.IncludeSprouts(relPath) && !args.ExcludeSprouts(relPath) {
						zlog.Info().
							Str("sprout", path).
							Msg("sprout run starting")
		
						err := dryad.StemRun(dryad.StemRunRequest{
							Garden: garden,
							StemPath:   path,
							Env:        env,
							Args:       args.Extras,
							JoinStdout: args.JoinStdout,
							JoinStderr: args.JoinStderr,
							Context:    args.Context,
						})
						if err != nil {
							zlog.Warn().
								Str("sprout", path).
								Err(err).
								Msg("sprout threw error during execution")
							if !args.IgnoreErrors {
								return err, nil
							}
						} else {
							zlog.Info().
								Str("sprout", path).
								Msg("sprout run finished")
						}
		
					}
		
					return nil, nil

				},
			},
		)
		if err != nil {
			zlog.Fatal().Err(err).Msg("error while crawling sprouts")
			return err, nil
		}

		return nil, nil
	}

	runSprouts = task.WithContext(
		runSprouts,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runSprouts,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while running sprouts")
				return 1
			}

			return 0
		},
	)

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
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

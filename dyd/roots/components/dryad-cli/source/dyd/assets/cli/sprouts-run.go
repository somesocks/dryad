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
		SproutFilter func (*task.ExecutionContext, *dryad.SafeSproutReference) (error, bool)
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

			err, includeExcludeFilter := dryad.SproutFilterFromCel(
				dryad.SproutFilterFromCelRequest{
					Include: includeOpts,
					Exclude: excludeOpts,
				},
			)
			if err != nil {
				return err, ParsedArgs{}
			}

			err, fromStdinFilter := ArgSproutFilterFromStdin(task.SERIAL_CONTEXT, req)
			if err != nil {
				return err, ParsedArgs{}
			}

			extras := args[0:]

			return nil, ParsedArgs{
				GardenPath: "",
				Parallel: parallel,
				SproutFilter: dryad.SproutFiltersCompose(
					fromStdinFilter,
					includeExcludeFilter,
				),
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

		err, sprouts := garden.Sprouts().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		// if confirm is set, we want to print the list
		// of sprouts to run
		if args.Confirm != "" {
			fmt.Println("dryad sprouts exec will execute these sprouts:")

			err := sprouts.Walk(
				ctx,
				dryad.SproutsWalkRequest{
					OnSprout: func (ctx *task.ExecutionContext, sprout *dryad.SafeSproutReference) (error, any) {
						
						err, shouldMatch := args.SproutFilter(ctx, sprout)
						if err != nil {
							return err, nil
						}

						if !shouldMatch {
							return nil, nil
						}
						
						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(sprout.Sprouts.Garden.BasePath, sprout.BasePath)
						if err != nil {
							return err, nil
						}	

						fmt.Println(" - " + relPath)

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

		err = sprouts.Walk(
			ctx,
			dryad.SproutsWalkRequest{
				OnSprout: func (ctx *task.ExecutionContext, sprout *dryad.SafeSproutReference) (error, any) {

					err, shouldMatch := args.SproutFilter(ctx, sprout)
					if err != nil {
						return err, nil
					}

					if shouldMatch {
						zlog.Info().
							Str("sprout", sprout.BasePath).
							Msg("sprout run starting")
		
						err := sprout.Run(
							ctx,
							dryad.SproutRunRequest{
								Env:        env,
								Args:       args.Extras,
								JoinStdout: args.JoinStdout,
								JoinStderr: args.JoinStderr,
								Context:    args.Context,
							},
						)
						if err != nil {
							zlog.Warn().
								Str("sprout", sprout.BasePath).
								Err(err).
								Msg("sprout threw error during execution")
							if !args.IgnoreErrors {
								return err, nil
							}
						} else {
							zlog.Info().
								Str("sprout", sprout.BasePath).
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
		WithOption(
			clib.NewOption(
				"from-stdin", 
				"if set, read a list of sprouts from stdin to use as a base list to run, instead of all sprouts. include and exclude filters are applied to this list. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(clib.NewOption("include", "choose which sprouts are included. the include filter is a CEL expression with access to a 'sprout' object that can be used to filter on properties of each sprout.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded.  the exclude filter is a CEL expression with access to a 'sprout' object that can be used to filter on properties of each sprout.").WithType(clib.OptionTypeMultiString)).
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

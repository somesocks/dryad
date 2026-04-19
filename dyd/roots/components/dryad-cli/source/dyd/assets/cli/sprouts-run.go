package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var sproutsRunCommand = func() clib.Command {
	type sproutsRunTarget struct {
		Sprout            *dryad.SafeSproutReference
		SproutRef         string
		VariantDescriptor string
	}

	type ParsedArgs struct {
		GardenPath           string
		Parallel             int
		Request              clib.ActionRequest
		IncludeExcludeFilter func(*task.ExecutionContext, *dryad.SafeSproutReference) (error, bool)
		Confirm              string
		Context              string
		VariantDescriptor    string
		Inherit              bool
		IgnoreErrors         bool
		JoinStdout           bool
		JoinStderr           bool
		LogStdout            string
		LogStderr            string
		Extras               []string
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
			var variantDescriptor string
			var inherit bool
			var ignoreErrors bool
			var confirm string
			var joinStdout bool
			var joinStderr bool
			var logStdout string
			var logStderr string

			var parallel int

			if options["variant"] != nil {
				variantDescriptor = options["variant"].(string)
			}

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

			if options["log-stdout"] != nil {
				logStdout = options["log-stdout"].(string)
				joinStdout = false
			}

			if options["log-stderr"] != nil {
				logStderr = options["log-stderr"].(string)
				joinStderr = false
			}

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			if options["confirm"] != nil {
				confirm = options["confirm"].(string)
			}

			err, includeExcludeFilter := dryad.SproutSelectorFilter(
				dryad.SelectorFilterRequest{
					Include: includeOpts,
					Exclude: excludeOpts,
				},
			)
			if err != nil {
				return err, ParsedArgs{}
			}

			extras := args[0:]

			return nil, ParsedArgs{
				GardenPath:           "",
				Parallel:             parallel,
				Request:              req,
				IncludeExcludeFilter: includeExcludeFilter,
				Confirm:              confirm,
				Context:              context,
				VariantDescriptor:    variantDescriptor,
				Inherit:              inherit,
				IgnoreErrors:         ignoreErrors,
				JoinStdout:           joinStdout,
				JoinStderr:           joinStderr,
				LogStdout:            logStdout,
				LogStderr:            logStderr,
				Extras:               extras,
			}
		},
	)

	var sproutsRunTargetsFromStdin = func(
		ctx *task.ExecutionContext,
		args ParsedArgs,
		sprouts *dryad.SafeSproutsReference,
	) (error, bool, []sproutsRunTarget) {
		var options = args.Request.Opts

		var fromStdin bool
		if options["from-stdin"] != nil {
			fromStdin = options["from-stdin"].(bool)
		}

		if !fromStdin {
			return nil, false, nil
		}

		targets := []sproutsRunTarget{}
		targetSet := map[string]bool{}
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			rawRef := strings.TrimSpace(scanner.Text())
			if rawRef == "" {
				continue
			}

			err, sproutRef := parseRootRef(rawRef)
			if err != nil {
				return err, false, nil
			}

			if args.VariantDescriptor != "" && sproutRef.HasSelector {
				return fmt.Errorf("sprouts run selector specified in both stdin sprout_ref and --variant"), false, nil
			}

			err, sproutPath := resolveSproutPath(sproutRef.Path)
			if err != nil {
				return err, false, nil
			}

			err, sprout := sprouts.Sprout(sproutPath).Resolve(ctx)
			if err != nil {
				return err, false, nil
			}

			variantDescriptor := args.VariantDescriptor
			if sproutRef.HasSelector {
				err, variantDescriptor = (dryad.RootVariantContext{Descriptor: sproutRef.Selector}).Filesystem()
				if err != nil {
					return err, false, nil
				}
			}

			targetKey := sprout.BasePath + "\x00" + variantDescriptor
			if targetSet[targetKey] {
				continue
			}
			targetSet[targetKey] = true

			targets = append(targets, sproutsRunTarget{
				Sprout:            sprout,
				SproutRef:         rawRef,
				VariantDescriptor: variantDescriptor,
			})
		}

		if err := scanner.Err(); err != nil {
			return err, false, nil
		}

		return nil, true, targets
	}

	var runSprouts = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, sprouts := garden.Sprouts().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, fromStdin, stdinTargets := sproutsRunTargetsFromStdin(ctx, args, sprouts)
		if err != nil {
			return err, nil
		}

		// if confirm is set, we want to print the list
		// of sprouts to run
		if args.Confirm != "" {
			fmt.Println("dryad sprouts exec will execute these sprouts:")

			if fromStdin {
				for _, target := range stdinTargets {
					err, shouldMatch := args.IncludeExcludeFilter(ctx, target.Sprout)
					if err != nil {
						return err, nil
					}
					if !shouldMatch {
						continue
					}

					fmt.Println(" - " + target.SproutRef)
				}
			} else {
				err := sprouts.Walk(
					ctx,
					dryad.SproutsWalkRequest{
						OnSprout: func(ctx *task.ExecutionContext, sprout *dryad.SafeSproutReference) (error, any) {
							err, shouldMatch := args.IncludeExcludeFilter(ctx, sprout)
							if err != nil {
								return err, nil
							}

							if !shouldMatch {
								return nil, nil
							}

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
					zlog.Error().Err(err).Msg("error while crawling sprouts")
					return err, nil
				}
			}

			fmt.Println("are you sure? type '" + args.Confirm + "' to continue")

			reader := bufio.NewReader(os.Stdin)

			input, err := reader.ReadString('\n')
			if err != nil {
				zlog.Error().Err(err).Msg("error while reading input")
				return err, nil
			}

			input = strings.TrimSuffix(input, "\n")

			if input != args.Confirm {
				zlog.Error().Msg("input does not match confirmation, aborting")
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

		runTarget := func(ctx *task.ExecutionContext, sprout *dryad.SafeSproutReference, variantDescriptor string) error {
			variantSelectorLabel := variantDescriptor
			if variantSelectorLabel == "" {
				variantSelectorLabel = "default"
			}

			zlog.Info().
				Str("sprout", sprout.BasePath).
				Str("variant_selector", variantSelectorLabel).
				Msg("sprout run requested")

			err := sprout.Run(
				ctx,
				dryad.SproutRunRequest{
					VariantDescriptor: variantDescriptor,
					Env:               env,
					Args:              args.Extras,
					JoinStdout:        args.JoinStdout,
					JoinStderr:        args.JoinStderr,
					LogStdout: struct {
						Path string
						Name string
					}{
						Path: args.LogStdout,
						Name: "",
					},
					LogStderr: struct {
						Path string
						Name string
					}{
						Path: args.LogStderr,
						Name: "",
					},
					Context: args.Context,
				},
			)
			if err != nil {
				zlog.Warn().
					Str("sprout", sprout.BasePath).
					Str("variant_selector", variantSelectorLabel).
					Err(err).
					Msg("sprout threw error during execution")
				if !args.IgnoreErrors {
					return err
				}
			} else {
				zlog.Info().
					Str("sprout", sprout.BasePath).
					Str("variant_selector", variantSelectorLabel).
					Msg("sprout run completed")
			}

			return nil
		}

		if fromStdin {
			for _, target := range stdinTargets {
				err, shouldMatch := args.IncludeExcludeFilter(ctx, target.Sprout)
				if err != nil {
					return err, nil
				}

				if !shouldMatch {
					continue
				}

				err = runTarget(ctx, target.Sprout, target.VariantDescriptor)
				if err != nil {
					return err, nil
				}
			}

			return nil, nil
		}

		err = sprouts.Walk(
			ctx,
			dryad.SproutsWalkRequest{
				OnSprout: func(ctx *task.ExecutionContext, sprout *dryad.SafeSproutReference) (error, any) {
					err, shouldMatch := args.IncludeExcludeFilter(ctx, sprout)
					if err != nil {
						return err, nil
					}

					if shouldMatch {
						err = runTarget(ctx, sprout, args.VariantDescriptor)
						if err != nil {
							return err, nil
						}
					}

					return nil, nil

				},
			},
		)
		if err != nil {
			zlog.Error().Err(err).Msg("error while crawling sprouts")
			return err, nil
		}

		return nil, nil
	}

	runSprouts = task.WithContext(
		runSprouts,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runSprouts,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while running sprouts")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("run", "run each sprout in the current garden").
		WithOption(
			clib.NewOption(
				"from-stdin",
				"if set, read a list of sprout refs from stdin to use as a base list to run, instead of all sprouts. include and exclude filters are applied to this list. default false",
			).
				WithType(clib.OptionTypeBool),
		).
		WithOption(clib.NewOption("include", "choose which sprouts are included. filter format: <path-glob>[~<selector>] or ~<selector>; selector keys match variants first, then traits.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which sprouts are excluded. filter format: <path-glob>[~<selector>] or ~<selector>; selector keys match variants first, then traits.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(clib.NewOption("variant", "variant descriptor selector for sprout stems (filesystem form: dimension=option+dimension=option). supports none/any/host; inherit is invalid for sprout runs")).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("confirm", "ask for a confirmation string to be entered to execute this command").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("ignore-errors", "continue running even if a sprout returns an error").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("join-stdout", "join the stdout of child processes to the stderr of the parent dryad process. default false").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("join-stderr", "join the stderr of child processes to the stderr of the parent dryad process. default false").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("log-stdout", "log the stdout of child processes to a file in the specified directory. disables joining").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("log-stderr", "log the stderr of child processes to a file in the specified directory. disables joining").WithType(clib.OptionTypeString)).
		WithArg(clib.NewArg("-- args", "args to pass to each sprout on execution").AsOptional()).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

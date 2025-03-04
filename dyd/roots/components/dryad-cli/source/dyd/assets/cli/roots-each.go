package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	// "fmt"

	"os"
	"os/exec"

	"strings"


	zlog "github.com/rs/zerolog/log"
)

var rootsEachCommand = func() clib.Command {

	type ParsedArgs struct {
		FromStdinFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
		IncludeExcludeFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)
		Shell string
		Command string
		GardenPath string
		Parallel int
		JoinStdout bool
		JoinStderr bool
		IgnoreErrors bool
	}

	var parseArgs =
		func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

			var path string = ""

			var joinStdout bool
			var joinStderr bool
	
			var ignoreErrors bool

			var includeOpts []string
			var excludeOpts []string

			if options["exclude"] != nil {
				excludeOpts = options["exclude"].([]string)
			}

			if options["include"] != nil {
				includeOpts = options["include"].([]string)
			}

			err, rootFilter := dryad.RootCelFilter(
				dryad.RootCelFilterRequest{
					Include: includeOpts,
					Exclude: excludeOpts,
				},
			)
			if err != nil {
				return err, ParsedArgs{}
			}

			err, fromStdinFilter := ArgRootFilterFromStdin(ctx, req)
			if err != nil {
				return err, ParsedArgs{}
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}
	
			var shell string
			if options["shell"] != nil {
				shell = options["shell"].(string)
			}
			if shell == "" {
				shell = os.Getenv("SHELL")
			}
			if shell == "" {
				shell = "/bin/sh"
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

			if options["ignore-errors"] != nil {
				ignoreErrors = options["ignore-errors"].(bool)
			} else {
				ignoreErrors = false
			}

			var command string = strings.Join(args[:], " ")

			return nil, ParsedArgs{
				GardenPath: path,
				Parallel: parallel,
				Shell: shell,
				Command: command,
				FromStdinFilter: fromStdinFilter,
				IncludeExcludeFilter: rootFilter,
				JoinStdout: joinStdout,
				JoinStderr: joinStderr,
				IgnoreErrors: ignoreErrors,
			}
		}

	var eachRoots = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
		if err != nil {
			return err, nil
		}

		err = roots.Walk(
			ctx,
			dryad.RootsWalkRequest{
				ShouldMatch: dryad.RootFiltersCompose(
					args.FromStdinFilter,
					args.IncludeExcludeFilter,
				),
				OnMatch: func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, any) {
					var err error

					cmd := exec.Command(
						args.Shell,
						[]string{"-c", args.Command}...,
					)
				
					cmd.Dir = root.BasePath

					cmd.Stdin = os.Stdin

					// optionally pipe the exec logs to us
					if args.JoinStdout {
						cmd.Stdout = os.Stdout
					}
				
					// optionally pipe the exec stderr to us
					if args.JoinStderr {
						cmd.Stderr = os.Stderr
					}
										
					err = cmd.Run()
					if err != nil && !args.IgnoreErrors {
						return err, nil
					}

					return nil, nil
				},
			},
		)
		return err, nil
	}

	eachRoots = task.WithContext(
		eachRoots,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)


	var action = task.Return(
		task.Series2(
			parseArgs,
			eachRoots,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while executing commands")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("each", "run a command once for each root. the current working directory is set to the base path of the root for each execution.").
		WithOption(clib.NewOption("include", "choose which roots are included in the list. the include filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("exclude", "choose which roots are excluded from the list.  the exclude filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.").WithType(clib.OptionTypeMultiString)).
		WithOption(
			clib.NewOption(
				"from-stdin", 
				"if set, read a list of roots from stdin to use as a base list of roots instead of all roots. include and exclude filters will be applied after this list. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.NewOption(
				"ignore-errors",
				"continue running even if one command execution returns an error",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.NewOption(
				"join-stdout",
				"join the stdout of child processes to the stderr of the parent dryad process. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.NewOption(
				"join-stderr",
				"join the stderr of child processes to the stderr of the parent dryad process. default false",
			).
			WithType(clib.OptionTypeBool),
		).
		WithOption(
			clib.NewOption(
				"shell",
				"override the shell used to run the command in each root",
			).
			WithType(clib.OptionTypeString),
		).
		WithArg(clib.NewArg("-- command", "command to run once for each root").AsOptional()).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()

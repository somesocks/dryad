package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	// "fmt"
	"path/filepath"

	"bufio"
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


	var buildStdinFilter = func (
		ctx *task.ExecutionContext,
		req clib.ActionRequest,
	) (error, func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)) {
		var options = req.Opts

		var fromStdin bool
		var fromStdinFilter func (*task.ExecutionContext, *dryad.SafeRootReference) (error, bool)

		var path = ""

		if options["from-stdin"] != nil {
			fromStdin = options["from-stdin"].(bool)
		} else {
			fromStdin = false
		}

		if fromStdin {
			unsafeGarden := dryad.Garden(path)
	
			err, garden := unsafeGarden.Resolve(ctx)
			if err != nil {
				return err, fromStdinFilter
			}
	
			err, roots := garden.Roots().Resolve(ctx)
			if err != nil {
				return err, fromStdinFilter
			}

			var rootSet = make(map[string]bool)
			var scanner = bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				var path = scanner.Text()
				var err error 
				var root dryad.SafeRootReference

				path, err = filepath.Abs(path)
				if err != nil {
					zlog.Error().
						Err(err).
						Msg("error reading path from stdin")
					return err, fromStdinFilter
				}

				path = _rootsOwningDependencyCorrection(path)
				err, root = roots.Root(path).Resolve(ctx)
				if err != nil {
					zlog.Error().
						Str("path", path).
						Err(err).
						Msg("error resolving root from path")
					return err, fromStdinFilter
				}

				rootSet[root.BasePath] = true
			}

			// Check for any errors during scanning
			if err := scanner.Err(); err != nil {
				zlog.Error().Err(err).Msg("error reading stdin")
				return err, fromStdinFilter
			}

			fromStdinFilter = func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, bool) {
				_, ok := rootSet[root.BasePath]
				return nil, ok
			}

		} else {
			fromStdinFilter = func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, bool) {
				return nil, true
			}
		}

		return nil, fromStdinFilter
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
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

			err, fromStdinFilter := buildStdinFilter(task.SERIAL_CONTEXT, req)
			if err != nil {
				return err, ParsedArgs{}
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
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
		},
	)

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
				OnMatch: func (ctx *task.ExecutionContext, root *dryad.SafeRootReference) (error, any) {
					var err error
					var shouldMatch bool

					err, shouldMatch = args.FromStdinFilter(ctx, root)
					if err != nil {
						return err, nil
					} else if !shouldMatch {
						return nil, nil
					}

					err, shouldMatch = args.IncludeExcludeFilter(ctx, root)
					if err != nil {
						return err, nil
					}

					if shouldMatch {

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

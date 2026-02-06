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
			type ParsedArgs struct {
				Path       string
				Extras     []string
				Context    string
				Override   string
				Inherit    bool
				Confirm    string
				JoinStdout bool
				JoinStderr bool
				LogStdout  string
				LogStderr  string
				Parallel   int
			}

			var parseArgs = task.From(
				func(req clib.ActionRequest) (error, ParsedArgs) {
					var args = req.Args
					var opts = req.Opts
					var context string
					var override string
					var inherit bool
					var confirm string
					var joinStdout bool
					var joinStderr bool
					var logStdout string
					var logStderr string
					var parallel int

					if opts["context"] != nil {
						context = opts["context"].(string)
					}

					if opts["inherit"] != nil {
						inherit = opts["inherit"].(bool)
					}

					if opts["override"] != nil {
						override = opts["override"].(string)
					}

					if opts["confirm"] != nil {
						confirm = opts["confirm"].(string)
					}

					if opts["join-stdout"] != nil {
						joinStdout = opts["join-stdout"].(bool)
					} else {
						joinStdout = true
					}

					if opts["join-stderr"] != nil {
						joinStderr = opts["join-stderr"].(bool)
					} else {
						joinStderr = true
					}

					if opts["log-stdout"] != nil {
						logStdout = opts["log-stdout"].(string)
						joinStdout = false
					}

					if opts["log-stderr"] != nil {
						logStderr = opts["log-stderr"].(string)
						joinStderr = false
					}

					if opts["parallel"] != nil {
						parallel = int(opts["parallel"].(int64))
					} else {
						parallel = PARALLEL_COUNT_DEFAULT
					}

					return nil, ParsedArgs{
						Path:       args[0],
						Extras:     args[1:],
						Context:    context,
						Override:   override,
						Inherit:    inherit,
						Confirm:    confirm,
						JoinStdout: joinStdout,
						JoinStderr: joinStderr,
						LogStdout:  logStdout,
						LogStderr:  logStderr,
						Parallel:   parallel,
					}
				},
			)

			var runSprout = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
				// if confirm is set, we want to print the list
				// of sprouts to run
				if args.Confirm != "" {
					fmt.Println("this package will be executed:")
					fmt.Println(args.Path)
					fmt.Println("are you sure? type '" + args.Confirm + "' to continue")

					reader := bufio.NewReader(os.Stdin)

					input, err := reader.ReadString('\n')
					if err != nil {
						return err, nil
					}

					input = strings.TrimSuffix(input, "\n")

					if input != args.Confirm {
						return fmt.Errorf("input does not match confirmation, aborting"), nil
					}
				}

				var env = map[string]string{}

				// pull
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

				err, path := resolveSproutPath(args.Path)
				if err != nil {
					return err, nil
				}

				err, garden := dryad.Garden(path).Resolve(ctx)
				if err != nil {
					return err, nil
				}

				err, sprouts := garden.Sprouts().Resolve(ctx)
				if err != nil {
					return err, nil
				}

				err, sprout := sprouts.Sprout(path).Resolve(ctx)
				if err != nil {
					return err, nil
				}

				err = sprout.Run(
					ctx,
					dryad.SproutRunRequest{
						MainOverride: args.Override,
						Env:          env,
						Args:         args.Extras,
						JoinStdout:   args.JoinStdout,
						LogStdout: struct {
							Path string
							Name string
						}{
							Path: args.LogStdout,
							Name: "",
						},
						JoinStderr: args.JoinStderr,
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
					return err, nil
				}

				return nil, nil
			}

			runSprout = task.WithContext(
				runSprout,
				func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
					return nil, task.NewContext(args.Parallel)
				},
			)

			return task.Return(
				task.Series2(
					parseArgs,
					runSprout,
				),
				func(err error, val any) int {
					if err != nil {
						zlog.Fatal().Err(err).Msg("error executing stem")
						return 1
					}
					return 0
				},
			)(req)
		})

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

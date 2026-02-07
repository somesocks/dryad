package cli

import (
	"bufio"
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var stemRunCommand = func() clib.Command {
	type ParsedArgs struct {
		Path     string
		Context  string
		Command  string
		Inherit  bool
		Confirm  string
		Extras   []string
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var context string
		var command string
		var inherit bool
		var confirm string
		var parallel int

		if options["context"] != nil {
			context = options["context"].(string)
		}

		if options["inherit"] != nil {
			inherit = options["inherit"].(bool)
		}

		if options["command"] != nil {
			command = options["command"].(string)
		}

		if options["confirm"] != nil {
			confirm = options["confirm"].(string)
		}

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Path:     args[0],
			Context:  context,
			Command:  command,
			Inherit:  inherit,
			Confirm:  confirm,
			Extras:   args[1:],
			Parallel: parallel,
		}
	}

	var runStem = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
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

		if args.Inherit {
			for _, e := range os.Environ() {
				if i := strings.Index(e, "="); i >= 0 {
					env[e[:i]] = e[i+1:]
				}
			}
		} else {
			env["TERM"] = os.Getenv("TERM")
		}

		err, garden := dryad.Garden(args.Path).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = dryad.StemRun(dryad.StemRunRequest{
			Garden:       garden,
			StemPath:     args.Path,
			MainOverride: args.Command,
			Env:          env,
			Args:         args.Extras,
			JoinStdout:   true,
			Context:      args.Context,
		})
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	runStem = task.WithContext(
		runStem,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			runStem,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error executing stem")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("run", "execute the main for a stem").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("command", "run this command in the stem run environment instead of the main")).
		WithOption(clib.NewOption("confirm", "ask for a confirmation string to be entered to execute this command").WithType(clib.OptionTypeString)).
		WithArg(clib.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

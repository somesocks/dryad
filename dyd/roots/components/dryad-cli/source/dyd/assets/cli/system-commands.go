package cli

import (
	clib "dryad/cli-builder"
	"dryad/task"
	"fmt"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func _appCommandsWalk(app clib.App, walk func(commands ...clib.Command)) {
	walk()

	var commands = app.Commands()
	for _, command := range commands {
		_commandCommandsWalk(walk, command)
	}
}

func _commandCommandsWalk(walk func(commands ...clib.Command), commands ...clib.Command) {
	walk(commands...)

	var lastCommand = commands[len(commands)-1]

	var subCommands = lastCommand.Commands()
	for _, subCommand := range subCommands {
		_commandCommandsWalk(walk, append(commands, subCommand)...)
	}

}

var systemCommands = func() clib.Command {
	type ParsedArgs struct {
		Parallel int
		App      clib.App
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var opts = req.Opts
		var parallel int

		if opts["parallel"] != nil {
			parallel = int(opts["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		return nil, ParsedArgs{
			Parallel: parallel,
			App:      req.App,
		}
	}

	var listCommands = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		_appCommandsWalk(args.App, func(commands ...clib.Command) {
			commandList := []string{"dryad"}

			for _, command := range commands {
				commandList = append(commandList, command.Key())
			}

			command := strings.Join(commandList, " ")

			fmt.Println(command)
		})

		return nil, nil
	}

	listCommands = task.WithContext(
		listCommands,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			listCommands,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while listing commands")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("commands", "print out a list of all dryad commands").
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()

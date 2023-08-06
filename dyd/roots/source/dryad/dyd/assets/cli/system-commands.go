package cli

import (
	clib "dryad/cli-builder"
	"fmt"
	"strings"
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
	command := clib.NewCommand("commands", "print out a list of all dryad commands").
		WithAction(func(req clib.ActionRequest) int {
			var app = req.App

			_appCommandsWalk(app, func(commands ...clib.Command) {
				commandList := []string{"dryad"}

				for _, command := range commands {
					commandList = append(commandList, command.Key())
				}

				command := strings.Join(commandList, " ")

				fmt.Println(command)
			})

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()

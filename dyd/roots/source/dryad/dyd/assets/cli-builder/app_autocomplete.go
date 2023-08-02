// Copyright (c) 2017. Oleg Sklyar & teris.io. All rights reserved.
// See the LICENSE file in the project root for licensing information.

package cli_builder

import (
	"strings"
)

func AppAutoComplete(app App, tokens []string) []string {
	var results = []string{}
	var commands = app.Commands()
	for _, command := range commands {
		results = append(results, CommandAutoComplete(command, tokens)...)
	}

	return results
}

func CommandAutoComplete(cmd Command, tokens []string) []string {
	var key = cmd.Key()
	var results = []string{}
	// fmt.Println("CommandAutoComplete", key, tokens)

	switch len(tokens) {
	case 0:
		{
			results = append(results, key+" ")
		}
	case 1:
		{
			if strings.HasPrefix(key, tokens[0]) {
				results = append(results, key+" ")
			}
		}
	default:
		{
			if key == tokens[0] {
				var subTokens = tokens[1:]
				var subCommands = cmd.Commands()
				if len(subCommands) > 0 {
					for _, subCommand := range subCommands {
						results = append(results, CommandAutoComplete(subCommand, subTokens)...)
					}
				} else {
					results = append(results, ArgumentsAutoComplete(
						cmd.Args(),
						cmd.Options(),
						subTokens,
					)...)
				}

			}
		}
	}

	return results
}

func ArgumentsAutoComplete(args []Arg, options []Option, tokens []string) []string {
	var results = []string{}
	// fmt.Println("ArgumentsAutoComplete", tokens, args, options)

	switch len(tokens) {
	case 0:
		{
			return results
		}
	case 1:
		{
			var token = tokens[0]
			// fmt.Println("case 1", token, strings.HasPrefix(token, "-"))
			if strings.HasPrefix(token, "-") {
				for _, option := range options {
					var optionKey = "--" + option.Key() + "="
					// fmt.Println("case 1 option", optionKey, strings.HasPrefix(optionKey, token))
					if strings.HasPrefix(optionKey, token) {
						results = append(results, optionKey)
					}
				}
			} else if len(args) > 0 {
				var arg = args[0]
				var ac = arg.AutoComplete()
				results = append(results, ac(token)...)
			}
		}
	default:
		{
			var token = tokens[0]
			if strings.HasPrefix(token, "-") {
				results = append(results, ArgumentsAutoComplete(args, options, tokens[1:])...)
			} else if len(args) > 0 {
				results = append(results, ArgumentsAutoComplete(args[1:], options, tokens[1:])...)
			}
		}
	}

	return results
}

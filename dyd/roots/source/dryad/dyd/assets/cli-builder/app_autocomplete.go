// Copyright (c) 2017. Oleg Sklyar & teris.io. All rights reserved.
// See the LICENSE file in the project root for licensing information.

package cli_builder

import (
	"strings"
)

func AppAutoComplete(app App, tokens []string) (error, []string) {
	var results = []string{}
	var commands = app.Commands()
	for _, command := range commands {
		var err, res = CommandAutoComplete(command, tokens)
		if err != nil {
			return err, results
		}
		results = append(results, res...)
	}

	return nil, results
}

func CommandAutoComplete(cmd Command, tokens []string) (error, []string) {
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
						var err, res = CommandAutoComplete(subCommand, subTokens)
						if err != nil {
							return err, results
						}
						results = append(results, res...)
					}
				} else {
					var err, res = ArgumentsAutoComplete(
						cmd.Args(),
						cmd.Options(),
						subTokens,
					)
					if err != nil {
						return err, results
					}
					results = append(results, res...)
				}

			}
		}
	}

	return nil, results
}

func ArgumentsAutoComplete(args []Arg, options []Option, tokens []string) (error, []string) {
	var results = []string{}
	// fmt.Println("ArgumentsAutoComplete", tokens, args, options)

	switch len(tokens) {
	case 0:
		{
			return nil, results
		}
	case 1:
		{
			var token = tokens[0]
			// fmt.Println("case 1", token, strings.HasPrefix(token, "-"))
			if strings.HasPrefix(token, "-") {
				for _, option := range options {
					switch option.Type() {
					case OptionTypeBool:
						var impliedKey = "--" + option.Key()
						if strings.HasPrefix(impliedKey, token) {
							results = append(results, impliedKey)
						}
						var equalsKey = impliedKey + "="
						if strings.HasPrefix(equalsKey, token) {
							results = append(results, equalsKey)
						}
					case OptionTypeMultiBool:
						var impliedKey = "--" + option.Key()
						if strings.HasPrefix(impliedKey, token) {
							results = append(results, impliedKey)
						}
						var equalsKey = impliedKey + "="
						if strings.HasPrefix(equalsKey, token) {
							results = append(results, equalsKey)
						}
					default:
						var optionKey = "--" + option.Key() + "="
						if strings.HasPrefix(optionKey, token) {
							results = append(results, optionKey)
						}
					}
				}
			} else if len(args) > 0 {
				var arg = args[0]
				var ac = arg.AutoComplete()
				var err, aac = ac(token)
				if err != nil {
					return err, results
				}
				results = append(results, aac...)
			}
		}
	default:
		{
			var token = tokens[0]
			if strings.HasPrefix(token, "-") {
				var err, res = ArgumentsAutoComplete(args, options, tokens[1:])
				if err != nil {
					return err, results
				}
				results = append(results, res...)
			} else if len(args) > 0 {
				var err, res = ArgumentsAutoComplete(args[1:], options, tokens[1:])
				if err != nil {
					return err, results
				}
				results = append(results, res...)
			}
		}
	}

	return nil, results
}

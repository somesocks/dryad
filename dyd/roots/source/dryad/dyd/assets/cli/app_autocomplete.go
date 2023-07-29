// Copyright (c) 2017. Oleg Sklyar & teris.io. All rights reserved.
// See the LICENSE file in the project root for licensing information.

package cli

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
	// fmt.Println("autocomplete", key, tokens)

	switch len(tokens) {
	case 0:
		{
			results = append(results, key)
		}
	case 1:
		{
			if strings.HasPrefix(key, tokens[0]) {
				results = append(results, key)
			}
		}
	default:
		{
			if key == tokens[0] {
				var subCommands = cmd.Commands()
				var subTokens = tokens[1:]
				for _, subCommand := range subCommands {
					results = append(results, CommandAutoComplete(subCommand, subTokens)...)
				}
			}
		}
	}

	return results
}

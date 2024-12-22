package core

import (
	"regexp"
)

var _RE_SCOPE_SHELL_MATCH = regexp.MustCompile(`((?:"(?:[^"\\]|\\.)*")|(?:\S+))`)

// parse a scope setting as a string of shell args
func ScopeSettingParseShell(setting string) ([]string, error) {

	args := _RE_SCOPE_SHELL_MATCH.FindAllString(setting, -1)

	return args, nil
}

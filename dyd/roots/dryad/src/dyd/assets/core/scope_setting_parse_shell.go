package core

import (
	"regexp"
)

var _RE_SCOPE_SHELL_MATCH = regexp.MustCompile(`((?:"(?:[^"\\]|\\.)*")|(?:\S+))`)

// parse a scope setting as a string of shell args
func ScopeSettingParseShell(basePath string, scope string, setting string) ([]string, error) {
	value, err := ScopeSettingGet(basePath, scope, setting)
	if err != nil {
		return nil, err
	}

	args := _RE_SCOPE_SHELL_MATCH.FindAllString(value, -1)

	return args, nil
}

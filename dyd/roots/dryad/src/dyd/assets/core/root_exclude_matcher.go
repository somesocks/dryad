package core

import "strings"

var _DEFAULT_ROOT_EXCLUDE_MATCHER = func(path string) bool {
	return false
}

func RootExcludeMatcher(args []string) func(path string) bool {
	if len(args) == 0 {
		return _DEFAULT_ROOT_EXCLUDE_MATCHER
	}

	return func(path string) bool {
		for _, arg := range args {
			if strings.Contains(path, arg) {
				return true
			}
		}
		return false
	}
}

package core

import "strings"

var _DEFAULT_ROOT_INCLUDE_MATCHER = func(path string) bool {
	return true
}

func RootIncludeMatcher(args []string) func(path string) bool {
	if len(args) == 0 {
		return _DEFAULT_ROOT_INCLUDE_MATCHER
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

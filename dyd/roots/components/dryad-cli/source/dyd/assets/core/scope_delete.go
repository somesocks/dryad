package core

import (
	fs2 "dryad/filesystem"
)

func ScopeDelete(path string, scope string) error {
	var scopePath, err = ScopePath(path, scope)
	if err != nil {
		return err
	}

	err = fs2.RemoveAll(scopePath)
	return err
}

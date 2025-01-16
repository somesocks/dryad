package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
)

func ScopeDelete(path string, scope string) error {
	var scopePath, err = ScopePath(path, scope)
	if err != nil {
		return err
	}

	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, scopePath)
	return err
}

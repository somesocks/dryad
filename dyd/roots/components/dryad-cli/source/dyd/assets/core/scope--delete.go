package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
)

func ScopeDelete(garden *SafeGardenReference, scope string) error {
	var scopePath, err = ScopePath(garden, scope)
	if err != nil {
		return err
	}

	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, scopePath)
	return err
}

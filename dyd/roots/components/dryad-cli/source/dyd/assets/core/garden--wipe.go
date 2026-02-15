package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"path/filepath"
)

func (sg *SafeGardenReference) Wipe(ctx *task.ExecutionContext) error {
	var err error
	var gardenPath string = sg.BasePath

	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")
	err, _ = fs2.RemoveAll(ctx, sproutsPath)
	if err != nil {
		return err
	}

	derivationsPath := filepath.Join(gardenPath, "dyd", "heap", "derivations")
	err, _ = fs2.RemoveAll(ctx, derivationsPath)
	if err != nil {
		return err
	}

	stemsPath := filepath.Join(gardenPath, "dyd", "heap", "stems")
	err, _ = fs2.RemoveAll(ctx, stemsPath)
	if err != nil {
		return err
	}

	heapSproutsPath := filepath.Join(gardenPath, "dyd", "heap", "sprouts")
	err, _ = fs2.RemoveAll(ctx, heapSproutsPath)
	if err != nil {
		return err
	}

	filesPath := filepath.Join(gardenPath, "dyd", "heap", "files")
	err, _ = fs2.RemoveAll(ctx, filesPath)
	if err != nil {
		return err
	}

	secretsPath := filepath.Join(gardenPath, "dyd", "heap", "secrets")
	err, _ = fs2.RemoveAll(ctx, secretsPath)
	if err != nil {
		return err
	}

	contextsPath := filepath.Join(gardenPath, "dyd", "heap", "contexts")
	err, _ = fs2.RemoveAll(ctx, contextsPath)
	if err != nil {
		return err
	}

	return nil
}

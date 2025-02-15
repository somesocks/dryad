package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"
)

func (sg *SafeGardenReference) Wipe(ctx *task.ExecutionContext) (error) {
	var err error
	var gardenPath string = sg.BasePath

	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")
	err, _ = fs2.RemoveAll(ctx, sproutsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(sproutsPath, os.ModePerm)
	if err != nil {
		return err
	}

	derivationsPath := filepath.Join(gardenPath, "dyd", "heap", "derivations")
	err, _ = fs2.RemoveAll(ctx, derivationsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(derivationsPath, os.ModePerm)
	if err != nil {
		return err
	}

	stemsPath := filepath.Join(gardenPath, "dyd", "heap", "stems")
	err, _ = fs2.RemoveAll(ctx, stemsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(stemsPath, os.ModePerm)
	if err != nil {
		return err
	}

	filesPath := filepath.Join(gardenPath, "dyd", "heap", "files")
	err, _ = fs2.RemoveAll(ctx, filesPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filesPath, os.ModePerm)
	if err != nil {
		return err
	}

	secretsPath := filepath.Join(gardenPath, "dyd", "heap", "secrets")
	err, _ = fs2.RemoveAll(ctx, secretsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(secretsPath, os.ModePerm)
	if err != nil {
		return err
	}

	contextsPath := filepath.Join(gardenPath, "dyd", "heap", "contexts")
	err, _ = fs2.RemoveAll(ctx, contextsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(contextsPath, os.ModePerm)
	if err != nil {
		return err
	}

	
	var unsafeGardenRef = Garden(sg.BasePath)

	err, _ = unsafeGardenRef.Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

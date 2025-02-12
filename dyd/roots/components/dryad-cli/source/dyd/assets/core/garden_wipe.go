package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"
)

type GardenWipeRequest struct {
	Garden *SafeGardenReference
}

func GardenWipe(ctx *task.ExecutionContext, req GardenWipeRequest) (error, any) {
	var err error
	var gardenPath string = req.Garden.BasePath

	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")
	err, _ = fs2.RemoveAll(ctx, sproutsPath)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(sproutsPath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	derivationsPath := filepath.Join(gardenPath, "dyd", "heap", "derivations")
	err, _ = fs2.RemoveAll(ctx, derivationsPath)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(derivationsPath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	stemsPath := filepath.Join(gardenPath, "dyd", "heap", "stems")
	err, _ = fs2.RemoveAll(ctx, stemsPath)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(stemsPath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	filesPath := filepath.Join(gardenPath, "dyd", "heap", "files")
	err, _ = fs2.RemoveAll(ctx, filesPath)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(filesPath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	secretsPath := filepath.Join(gardenPath, "dyd", "heap", "secrets")
	err, _ = fs2.RemoveAll(ctx, secretsPath)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(secretsPath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	contextsPath := filepath.Join(gardenPath, "dyd", "heap", "contexts")
	err, _ = fs2.RemoveAll(ctx, contextsPath)
	if err != nil {
		return err, nil
	}
	err = os.MkdirAll(contextsPath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	err, _ = GardenCreate(ctx, GardenCreateRequest{BasePath: gardenPath})
	if err != nil {
		return err, nil
	}

	return nil, nil
}

package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"
)

func GardenWipe(gardenPath string) error {

	// normalize garden path
	gardenPath, err := GardenPath(gardenPath)
	if err != nil {
		return err
	}

	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")
	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, sproutsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(sproutsPath, os.ModePerm)
	if err != nil {
		return err
	}

	derivationsPath := filepath.Join(gardenPath, "dyd", "heap", "derivations")
	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, derivationsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(derivationsPath, os.ModePerm)
	if err != nil {
		return err
	}

	stemsPath := filepath.Join(gardenPath, "dyd", "heap", "stems")
	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, stemsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(stemsPath, os.ModePerm)
	if err != nil {
		return err
	}

	filesPath := filepath.Join(gardenPath, "dyd", "heap", "files")
	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, filesPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filesPath, os.ModePerm)
	if err != nil {
		return err
	}

	secretsPath := filepath.Join(gardenPath, "dyd", "heap", "secrets")
	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, secretsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(secretsPath, os.ModePerm)
	if err != nil {
		return err
	}

	contextsPath := filepath.Join(gardenPath, "dyd", "heap", "contexts")
	err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, contextsPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(contextsPath, os.ModePerm)
	if err != nil {
		return err
	}

	err, _ = GardenCreate(task.DEFAULT_CONTEXT, GardenCreateRequest{BasePath: gardenPath})
	if err != nil {
		return err
	}

	return nil
}

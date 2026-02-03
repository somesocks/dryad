package core

import "io/fs"

func SecretsExist(path string) (bool, error) {
	var err error
	var exists bool
	var found bool

	path, err = SecretsPath(path)
	if err != nil {
		return false, err
	}

	// return false if the folder doesn't exist
	exists, err = fileExists(path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	err = SecretsWalk(
		SecretsWalkArgs{
			BasePath: path,
			OnMatch: func(_ string, info fs.FileInfo) error {
				if info.IsDir() {
					return nil
				}
				found = true
				return nil
			},
		},
	)
	if err != nil {
		return false, err
	}

	return found, nil
}

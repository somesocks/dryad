package core

import (
	"path/filepath"
)

func SecretsPath(path string) (string, error) {
	var stemPath string
	var err error

	stemPath, err = StemPath(path)
	if err != nil {
		return "", err
	}

	var secretsPath = filepath.Join(stemPath, "dyd", "secrets")

	return secretsPath, nil
}

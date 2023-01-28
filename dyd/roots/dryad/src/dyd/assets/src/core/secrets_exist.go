package core

func SecretsExist(path string) (bool, error) {
	var err error
	var exists bool
	var fingerprint string

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

	// use the fingerprint as a proxy to check if there are actually any secrets,
	// or if the secrets folder is empty
	fingerprint, err = SecretsFingerprint(SecretsFingerprintArgs{BasePath: path})
	if err != nil {
		return false, err
	}

	return fingerprint != "", nil
}

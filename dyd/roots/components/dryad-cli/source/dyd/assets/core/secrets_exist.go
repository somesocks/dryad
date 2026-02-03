package core

func SecretsExist(path string) (bool, error) {
	secretsPath, err := SecretsPath(path)
	if err != nil {
		return false, err
	}
	return fileExists(secretsPath)
}

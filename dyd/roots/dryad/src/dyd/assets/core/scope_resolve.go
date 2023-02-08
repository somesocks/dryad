package core

func ScopeResolve(basePath string, command string) (string, error) {
	_, err := GardenPath(basePath)

	return "", err
}

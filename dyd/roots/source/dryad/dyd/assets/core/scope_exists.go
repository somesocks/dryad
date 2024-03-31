package core

func ScopeExists(basePath string, scope string) (bool, error) {
	scopePath, err := ScopePath(basePath, scope)
	// fmt.Println("[debug] scopePath", scopePath, err)
	if err != nil {
		return false, err
	}

	scopeExists, err := fileExists(scopePath)
	// fmt.Println("[debug] scopeExists", scopeExists, err)
	if err != nil {
		return false, err
	}

	return scopeExists, nil
}

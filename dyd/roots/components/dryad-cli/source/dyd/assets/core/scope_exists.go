package core

func ScopeExists(garden *SafeGardenReference, scope string) (bool, error) {
	scopePath, err := ScopePath(garden, scope)
	if err != nil {
		return false, err
	}

	scopeExists, err := fileExists(scopePath)
	if err != nil {
		return false, err
	}

	return scopeExists, nil
}

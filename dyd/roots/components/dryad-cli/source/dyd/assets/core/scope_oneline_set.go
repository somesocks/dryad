package core

func ScopeOnelineSet(basePath string, scope string, value string) error {
	return ScopeSettingSet(basePath, scope, ".oneline", value)
}

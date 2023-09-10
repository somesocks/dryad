package core

func ScopeOnelineGet(basePath string, scope string) (string, error) {
	return ScopeSettingGet(basePath, scope, ".oneline")
}

package core

func ScopeOnelineSet(garden *SafeGardenReference, scope string, value string) error {
	return ScopeSettingSet(garden, scope, ".oneline", value)
}

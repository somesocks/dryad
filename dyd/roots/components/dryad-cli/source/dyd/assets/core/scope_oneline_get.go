package core

func ScopeOnelineGet(garden *SafeGardenReference, scope string) (string, error) {
	return ScopeSettingGet(garden, scope, ".oneline")
}

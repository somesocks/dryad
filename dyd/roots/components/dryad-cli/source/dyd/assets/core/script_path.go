package core

import (
	"fmt"
)

type ScriptPathRequest struct {
	Garden *SafeGardenReference
	Scope    string
	Setting  string
}

func ScriptPath(request ScriptPathRequest) (string, error) {
	runPath, err := ScopeSettingPath(request.Garden, request.Scope, request.Setting)
	if err != nil {
		return "", err
	} else if runPath == "" {
		return "", fmt.Errorf("%s not found in scope %s", request.Setting, request.Scope)
	} else {
		return runPath, err
	}
}

package core

import (
	"fmt"
)

type ScriptPathRequest struct {
	BasePath string
	Scope    string
	Setting  string
}

func ScriptPath(request ScriptPathRequest) (string, error) {
	runPath, err := ScopeSettingPath(request.BasePath, request.Scope, request.Setting)
	if err != nil {
		return "", err
	} else if runPath == "" {
		return "", fmt.Errorf("%s not found in scope %s", request.Setting, request.Scope)
	} else {
		return runPath, err
	}
}

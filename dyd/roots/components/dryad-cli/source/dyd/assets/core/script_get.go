package core

import (
	"io/ioutil"
)

type ScriptGetRequest struct {
	BasePath string
	Scope    string
	Setting  string
}

func ScriptGet(request ScriptGetRequest) (string, error) {
	scriptPath, err := ScriptPath(ScriptPathRequest(request))
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}

	// Convert []byte to string and print to screen
	script := string(bytes)
	return script, nil
}

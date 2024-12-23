package core

import "io/ioutil"

func ScriptOnelineGet(scriptPath string) (string, error) {
	bytes, err := ioutil.ReadFile(scriptPath + ".oneline")
	if err != nil {
		return "", err
	}

	// Convert []byte to string and print to screen
	oneline := string(bytes)
	return oneline, nil
}

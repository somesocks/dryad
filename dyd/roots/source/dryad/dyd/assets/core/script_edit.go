package core

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type ScriptEditRequest struct {
	BasePath string
	Scope    string
	Setting  string
	Env      map[string]string
}

func ScriptEdit(request ScriptEditRequest) error {
	scriptPath, err := ScopeSettingPath(request.BasePath, request.Scope, request.Setting)
	if err != nil {
		return err
	}

	var editor string

	if request.Env["EDITOR"] != "" {
		editor = request.Env["EDITOR"]
	} else if request.Env["VISUAL"] != "" {
		editor = request.Env["VISUAL"]
	} else {
		return fmt.Errorf("no editor found")
	}

	cmd := exec.Command(
		editor,
		scriptPath,
	)

	// prepare env
	cmd.Env = []string{}
	for key, val := range request.Env {
		cmd.Env = append(cmd.Env, key+"="+val)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(
		cmd.Env,
		// "HOME="+contextPath,
		// "DYD_CONTEXT="+contextPath,
		// "DYD_STEM="+stemPath,
		// "DYD_GARDEN="+gardenPath,
		"DYD_OS="+runtime.GOOS,
		"DYD_ARCH="+runtime.GOARCH,
	)

	err = cmd.Run()
	return err
}

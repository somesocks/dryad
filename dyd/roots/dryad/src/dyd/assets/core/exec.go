package core

import (
	"os"
	"os/exec"
	"runtime"
)

type ExecRequest struct {
	BasePath string
	Scope    string
	Setting  string
	Env      map[string]string
	Args     []string
}

func Exec(request ExecRequest) error {
	runPath, err := ScopeSettingPath(request.BasePath, request.Scope, request.Setting)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		runPath,
		request.Args...,
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

package core

import (
	"os"
	"os/exec"
	"runtime"
)

func Exec(basePath string, scope string, setting string, args []string) error {
	runPath, err := ScopeSettingPath(basePath, scope, setting)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		runPath,
		args...,
	)

	// prepare env
	cmd.Env = []string{}
	// for key, val := range env {
	// 	cmd.Env = append(cmd.Env, key+"="+val)
	// }

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

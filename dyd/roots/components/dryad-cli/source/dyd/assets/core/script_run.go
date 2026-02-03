package core

import (
	"os"
	"os/exec"
	"runtime"

	zerolog "github.com/rs/zerolog"
)

type ScriptRunRequest struct {
	Garden *SafeGardenReference
	Scope      string
	Setting    string
	Env        map[string]string
	Args       []string
}

func ScriptRun(request ScriptRunRequest) error {
	runPath, err := ScopeSettingPath(request.Garden, request.Scope, request.Setting)
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
		"DYD_GARDEN="+request.Garden.BasePath,
		"DYD_OS="+runtime.GOOS,
		"DYD_ARCH="+runtime.GOARCH,
		"DYD_LOG_LEVEL="+zerolog.GlobalLevel().String(),
	)

	// Always override DOCKER_HOST because context HOME changes can invalidate user paths.
	if dockerSock := GetDockerSockPath(); dockerSock != "" {
		cmd.Env = append(cmd.Env, "DOCKER_HOST=unix://"+dockerSock)
	}

	err = cmd.Run()
	return err
}

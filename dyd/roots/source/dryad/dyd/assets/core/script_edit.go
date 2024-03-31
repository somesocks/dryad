package core

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	zerolog "github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

type ScriptEditRequest struct {
	BasePath string
	Scope    string
	Setting  string
	Env      map[string]string
}

func ScriptEdit(request ScriptEditRequest) error {

	// verify that the scope exists
	scopeExists, err := ScopeExists(request.BasePath, request.Scope)
	if err != nil {
		return err
	}

	if !scopeExists {
		return fmt.Errorf("scope %s does not exist", request.Scope)
	}

	scriptPath, err := ScopeSettingPath(request.BasePath, request.Scope, request.Setting)
	if err != nil {
		return err
	}

	zlog.
		Debug().
		Str("scriptPath", scriptPath).
		Msg("script path")

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(scriptPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		zlog.
			Fatal().
			Err(err).
			Msg("error creating script file")
		return err
	}
	if err := f.Close(); err != nil {
		zlog.
			Fatal().
			Err(err).
			Msg("error closing script file after creation")
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
		"DYD_LOG_LEVEL="+zerolog.GlobalLevel().String(),
	)

	err = cmd.Run()
	return err
}

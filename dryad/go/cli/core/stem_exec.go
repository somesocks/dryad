package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type StemExecRequest struct {
	StemPath   string
	ExecPath   string
	Env        map[string]string
	Args       []string
	JoinStdout bool
}

func StemExec(request StemExecRequest) error {
	var execPath = request.ExecPath
	var stemPath = request.StemPath
	var env = request.Env
	var args = request.Args

	if !filepath.IsAbs(stemPath) {
		if execPath != "" {
			stemPath = filepath.Join(filepath.Dir(execPath), stemPath)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			stemPath = filepath.Join(wd, stemPath)
		}
	}

	// prepare by getting the executable path
	dryadPath, err := os.Executable()
	if err != nil {
		return err
	}
	dryadPath, err = filepath.EvalSymlinks(dryadPath)
	if err != nil {
		return err
	}
	dryadPath = filepath.Dir(dryadPath)

	// rootMain := filepath.Join(finalStemPath, "dyd", "main")

	cmd := exec.Command(
		stemPath+"/dyd/main",
		args...,
	)

	cmd.Env = os.Environ()

	if len(env) > 0 {
		for key, val := range env {
			cmd.Env = append(cmd.Env, key+"="+val)
		}
	}

	cmd.Stdin = os.Stdin

	// optionally pipe the exec logs to us
	if request.JoinStdout {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr

	envPath := fmt.Sprintf(
		"PATH=%s:%s:%s",
		stemPath+"/dyd/path",
		dryadPath,
		"/usr/bin/",
	)

	cmd.Env = append(cmd.Env, envPath)

	cmd.Dir = stemPath

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

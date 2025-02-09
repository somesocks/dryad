package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	zerolog "github.com/rs/zerolog"
)

type StemRunRequest struct {
	Garden *SafeGardenReference
	StemPath     string
	WorkingPath  string
	MainOverride string
	Context      string
	Env          map[string]string
	Args         []string
	JoinStdout   bool
	JoinStderr   bool
	InheritEnv   bool
}

func stemRun_prepContext(request StemRunRequest) (string, error) {
	var gardenPath string
	var err error
	var context string

	context = request.Context
	if context == "" {
		context = "default"
	}

	gardenPath = request.Garden.BasePath

	contextPath := filepath.Join(gardenPath, "dyd", "heap", "contexts", context)
	err = os.MkdirAll(contextPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return contextPath, nil
}

func StemRun(request StemRunRequest) error {
	var workingPath = request.WorkingPath
	var stemPath = request.StemPath
	var env = request.Env
	var args = request.Args
	var gardenPath string
	var err error

	if env == nil {
		env = make(map[string]string)
	}

	if !filepath.IsAbs(stemPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		stemPath = filepath.Join(cwd, stemPath)
	}

	gardenPath = request.Garden.BasePath

	contextPath, err := stemRun_prepContext(request)
	if err != nil {
		return err
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

	var command string
	if request.MainOverride != "" {
		command = request.MainOverride
	} else {
		command = stemPath + "/dyd/commands/default"
	}

	cmd := exec.Command(
		command,
		args...,
	)

	cmd.Dir = workingPath

	// prepare env
	cmd.Env = []string{}

	if request.InheritEnv {
		cmd.Env = append(
			cmd.Env, os.Environ()...)
	}

	for key, val := range env {
		cmd.Env = append(cmd.Env, key+"="+val)
	}

	cmd.Stdin = os.Stdin

	// optionally pipe the exec logs to us
	if request.JoinStdout {
		cmd.Stdout = os.Stdout
	}

	// optionally pipe the exec stderr to us
	if request.JoinStderr {
		cmd.Stderr = os.Stderr
	}

	envPath := fmt.Sprintf(
		"PATH=%s/dyd/commands:%s/dyd/path:%s:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		stemPath,
		stemPath,
		dryadPath,
	)

	cmd.Env = append(
		cmd.Env,
		envPath,
		"HOME="+contextPath,
		"DYD_CONTEXT="+contextPath,
		"DYD_STEM="+stemPath,
		"DYD_GARDEN="+gardenPath,
		"DYD_OS="+runtime.GOOS,
		"DYD_ARCH="+runtime.GOARCH,
		"DYD_LOG_LEVEL="+zerolog.GlobalLevel().String(),
	)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

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
	StemPath     string
	WorkingPath  string
	MainOverride string
	GardenPath   string
	Context      string
	Env          map[string]string
	Args         []string
	JoinStdout   bool
	InheritEnv   bool
}

func stemRun_prepContext(request StemRunRequest) (string, error) {
	context := request.Context
	if context == "" {
		context = "default"
	}

	var gardenPath string
	var err error
	if request.GardenPath != "" {
		gardenPath = request.GardenPath
	} else {
		gardenPath, err = GardenPath(request.StemPath)
		if err != nil {
			return "", err
		}
	}

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

	var gardenPath string
	var err error
	if request.GardenPath != "" {
		gardenPath = request.GardenPath
	} else {
		gardenPath, err = GardenPath(request.StemPath)
		if err != nil {
			return err
		}
	}

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

	cmd.Stderr = os.Stderr
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

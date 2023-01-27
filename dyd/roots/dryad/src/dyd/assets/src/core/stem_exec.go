package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type StemExecRequest struct {
	StemPath   string
	ExecPath   string
	Context    string
	Env        map[string]string
	Args       []string
	JoinStdout bool
}

func stemExec_prepContext(request StemExecRequest) (string, error) {
	context := request.Context
	if context == "" {
		context = "default"
	}

	gardenPath, err := GardenPath(request.StemPath)
	if err != nil {
		return "", err
	}

	contextPath := filepath.Join(gardenPath, "dyd", "heap", "contexts", context)
	err = os.MkdirAll(contextPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return contextPath, nil
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

	gardenPath, err := GardenPath(request.StemPath)
	if err != nil {
		return err
	}

	contextPath, err := stemExec_prepContext(request)
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

	// rootMain := filepath.Join(finalStemPath, "dyd", "main")

	cmd := exec.Command(
		stemPath+"/dyd/main",
		args...,
	)

	// prepare env
	cmd.Env = []string{}
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
		"PATH=%s:%s:%s",
		stemPath+"/dyd/path",
		dryadPath,
		"/usr/bin/",
	)

	// set the working directory to be the base path to the stem
	cmd.Dir = stemPath

	cmd.Env = append(
		cmd.Env,
		envPath,
		"PWD="+stemPath,
		"HOME="+contextPath,
		"DYD_CONTEXT="+contextPath,
		"DYD_STEM="+stemPath,
		"DYD_GARDEN="+gardenPath,
		"DYD_OS="+runtime.GOOS,
		"DYD_ARCH="+runtime.GOARCH,
	)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

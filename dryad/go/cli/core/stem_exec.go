package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func StemExec(stemPath string, env map[string]string, args ...string) error {

	if !filepath.IsAbs(stemPath) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		stemPath = filepath.Join(wd, stemPath)
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

	if len(env) > 0 {
		var envList []string

		for key, val := range env {
			envList = append(envList, key+"="+val)
		}

		cmd.Env = envList
	} else {
		cmd.Env = []string{}
	}

	// pipe the exec logs to us
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	envPath := fmt.Sprintf(
		"PATH=%s:%s:%s",
		dryadPath,
		stemPath+"/dyd/path",
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

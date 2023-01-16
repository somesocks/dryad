package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func StemExec(stemPath string, args ...string) error {
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

	// pipe the exec logs to us
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	envPath := fmt.Sprintf(
		"PATH=%s:%s:%s",
		dryadPath,
		stemPath+"/dyd/path",
		"/usr/bin/",
	)

	cmd.Env = []string{
		envPath,
	}

	cmd.Dir = stemPath

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

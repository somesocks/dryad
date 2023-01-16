package core

import (
	"fmt"
	"os"
	"os/exec"
)

func StemExec(stemPath string, args ...string) error {
	// rootMain := filepath.Join(finalStemPath, "dyd", "main")

	cmd := exec.Command(
		stemPath + "/dyd/main",
	)

	// pipe the exec logs to us
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	envPath := fmt.Sprintf(
		"PATH=%s:%s",
		stemPath+"/dyd/path",
		"/usr/bin/",
	)

	cmd.Env = []string{
		envPath,
	}

	cmd.Dir = stemPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

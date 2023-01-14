package core

import "os/exec"

func StemExec(stemPath string, args ...string) error {
	// rootMain := filepath.Join(finalStemPath, "dyd", "main")

	var extendedArgs = []string{
		"-c",
		"cd " + stemPath + " && ./dyd/main $@",
		"dyd-main",
	}

	extendedArgs = append(extendedArgs, args...)

	cmd := exec.Command(
		"sh",
		extendedArgs...,
	)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

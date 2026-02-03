package core

import (
	"os"
	"path/filepath"

	"dryad/task"
)

// stemFinalize takes a partial stem and generates any files needed,
// such as fingerprints
func stemFinalize(ctx *task.ExecutionContext, stemPath string) (error, string) {
	var err error

	// normalize stemPath
	stemPath, err = StemPath(stemPath)
	if err != nil {
		return err, ""
	}

	// write the type file
	err = os.WriteFile(filepath.Join(stemPath, "dyd", "type"), []byte("stem"), os.ModePerm)
	if err != nil {
		return err, ""
	}

	// write out the stem fingerprint
	err, stemFingerprint := StemFingerprint(
		ctx,
		StemFingerprintRequest{
			BasePath: stemPath,
		},
	)
	if err != nil {
		return err, ""
	}

	err = os.WriteFile(filepath.Join(stemPath, "dyd", "fingerprint"), []byte(stemFingerprint), os.ModePerm)
	if err != nil {
		return err, ""
	}

	return nil, stemFingerprint
}

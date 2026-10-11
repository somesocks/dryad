package core

import (
	"context"
	"dryad/internal/filepath"
	"dryad/internal/os"
)

// sproutFinalize takes a partial sprout and generates package files,
// such as the sprout type sentinel and fingerprint.
func sproutFinalize(ctx context.Context, sproutPath string) (error, string) {
	var err error

	// write the type file
	err = os.WriteFile(filepath.Join(sproutPath, "dyd", "type"), []byte(SentinelSprout.String()), os.ModePerm)
	if err != nil {
		return err, ""
	}

	// write out the sprout fingerprint
	err, sproutFingerprint := StemFingerprint(
		ctx,
		StemFingerprintRequest{
			BasePath: sproutPath,
		},
	)
	if err != nil {
		return err, ""
	}

	err = os.WriteFile(filepath.Join(sproutPath, "dyd", "fingerprint"), []byte(sproutFingerprint), os.ModePerm)
	if err != nil {
		return err, ""
	}

	return nil, sproutFingerprint
}

package core

import (
	// "archive/tar"
	// "compress/gzip"
	// "errors"
	"io/fs"
	"os"
	"path/filepath"

	fs2 "dryad/filesystem"
)

type StemPackRequest struct {
	StemPath string
	TargetPath string
	IncludeDependencies bool
}

func stemPack(context BuildContext, request StemPackRequest) (string, error) {
	var stemPath = request.StemPath
	var targetPath = request.TargetPath
	var err error

	// convert relative stem path to absolute
	stemPath, err = filepath.Abs(stemPath) 
	if err != nil {
		return "", err
	}

	// resolve the dir to the root of the stem
	stemPath, err = StemPath(stemPath)
	if err != nil {
		return "", err
	}

	// convert relative target to absolute
	targetPath, err = filepath.Abs(targetPath) 
	if err != nil {
		return "", err
	}	

	stemFingerprint, err := _readFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return "", err
	}

	_, contextHasFingerprint := context.Fingerprints[stemPath]

	if contextHasFingerprint {
		return stemPath, nil
	}

	context.Fingerprints[stemPath] = stemFingerprint	

	// walk through the dependencies, and add them to the archive
	if request.IncludeDependencies {
		dependenciesPath := filepath.Join(stemPath, "dyd", "dependencies")

		dependencies, err := filepath.Glob(filepath.Join(dependenciesPath, "*"))
		if err != nil {
			return "", err
		}

		for _, dependencyPath := range dependencies {
			dependencyPath, err = filepath.EvalSymlinks(dependencyPath)
			if err != nil {
				return "", err
			}

			stemPack(context, StemPackRequest{
				StemPath: dependencyPath,
				TargetPath: targetPath,
				IncludeDependencies: request.IncludeDependencies,
			})
		}
	}

	var packedStemPath string
	packedStemPath, err = HeapAddStem(targetPath, stemPath)
	if err != nil {
		return "", err
	}

	sproutPath := filepath.Join(targetPath, "dyd", "sprouts", "main")
	sproutParent := filepath.Dir(sproutPath)
	sproutHeapPath := packedStemPath
	relSproutLink, err := filepath.Rel(
		sproutParent,
		sproutHeapPath,
	)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] building sprout parent")
	err = fs2.MkDir(sproutParent, fs.ModePerm)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] setting write permission on sprout parent")
	err = os.Chmod(sproutParent, 0o711)
	if err != nil {
		return "", err
	}

	tmpSproutPath := sproutPath + ".tmp"
	// fmt.Println("[debug] adding temporary sprout link")
	err = os.Symlink(relSproutLink, tmpSproutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] renaming sprout link", sproutPath)
	err = os.Rename(tmpSproutPath, sproutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] setting read permissions on sprout parent")
	err = os.Chmod(sproutParent, 0o511)
	if err != nil {
		return "", err
	}

	return sproutPath, nil
}

func StemPack(request StemPackRequest) (string, error) {
	var buildContext BuildContext = BuildContext{
		Fingerprints: map[string]string{},
	}

	err := os.MkdirAll(request.TargetPath, os.ModePerm)
	if err != nil {
		return "", err
	}	
	
	err = GardenCreate(request.TargetPath)
	if err != nil {
		return "", err
	}

	_, err = stemPack(buildContext, request)

	return request.TargetPath, err
}

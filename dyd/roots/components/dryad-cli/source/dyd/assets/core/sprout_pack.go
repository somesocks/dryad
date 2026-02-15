package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

type sproutPackRequest struct {
	SourceSproutPath string
	TargetGarden     *SafeGardenReference
}

func sproutPackPath(path string) (string, error) {
	zlog.Trace().
		Str("path", path).
		Msg("SproutPackPath")

	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	workingPath := path
	flagPath := filepath.Join(workingPath, "dyd", "type")
	fileBytes, fileErr := os.ReadFile(flagPath)

	for workingPath != "/" {
		if fileErr == nil {
			warnTypeFileWhitespace(flagPath, SentinelSprout.String(), string(fileBytes))

			sentinel := SentinelFromString(strings.TrimSpace(string(fileBytes)))
			if sentinel == SentinelSprout {
				return workingPath, nil
			}

			return "", errors.New("malformed type file, or sprout search started inside non-sprout resource")
		}

		workingPath = filepath.Dir(workingPath)
		flagPath = filepath.Join(workingPath, "dyd", "type")
		fileBytes, fileErr = os.ReadFile(flagPath)
	}

	return "", errors.New("dyd sprout path not found starting from " + path)
}

func sproutPack(
	ctx *task.ExecutionContext,
	context BuildContext,
	request sproutPackRequest,
) (string, error) {
	var err error
	var sproutPath = request.SourceSproutPath

	zlog.
		Debug().
		Str("sproutPath", sproutPath).
		Str("targetPath", request.TargetGarden.BasePath).
		Msg("SproutPack packing sprout")

	sproutPath, err = sproutPackPath(sproutPath)
	if err != nil {
		return "", err
	}

	dependenciesPath := filepath.Join(sproutPath, "dyd", "dependencies")
	dependencies, err := filepath.Glob(filepath.Join(dependenciesPath, "*"))
	if err != nil {
		return "", err
	}

	for _, dependencyPath := range dependencies {
		dependencyPath, err = filepath.EvalSymlinks(dependencyPath)
		if err != nil {
			return "", err
		}

		_, err = stemPack(
			ctx,
			context,
			stemPackRequest{
				SourceStemPath: dependencyPath,
				TargetGarden:   request.TargetGarden,
			},
		)
		if err != nil {
			return "", err
		}
	}

	err, targetHeap := request.TargetGarden.Heap().Resolve(ctx)
	if err != nil {
		return "", err
	}

	err, targetSprouts := targetHeap.Sprouts().Resolve(ctx)
	if err != nil {
		return "", err
	}

	err, packedSprout := targetSprouts.AddSprout(
		task.SERIAL_CONTEXT,
		HeapAddSproutRequest{
			SproutPath: sproutPath,
		},
	)
	if err != nil {
		return "", err
	}

	return packedSprout.BasePath, nil
}

type SproutPackRequest struct {
	SourceSproutPath string
	TargetPath       string
	Format           string
}

func SproutPack(
	ctx *task.ExecutionContext,
	request SproutPackRequest,
) (string, error) {
	zlog.
		Debug().
		Str("sproutPath", request.SourceSproutPath).
		Str("targetPath", request.TargetPath).
		Str("format", request.Format).
		Msg("SproutPack")

	var buildContext BuildContext = BuildContext{
		Fingerprints: map[string]string{},
	}

	err := os.MkdirAll(request.TargetPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	var unsafeTargetGardenRef = Garden(request.TargetPath)
	err, safeTargetGardenRef := unsafeTargetGardenRef.Create(ctx)
	if err != nil {
		return "", err
	}

	packedSproutPath, err := sproutPack(
		ctx,
		buildContext,
		sproutPackRequest{
			TargetGarden:     safeTargetGardenRef,
			SourceSproutPath: request.SourceSproutPath,
		},
	)
	if err != nil {
		return "", err
	}

	_, err = finalizeSproutPath(safeTargetGardenRef, packedSproutPath)
	if err != nil {
		return "", err
	}

	return stemArchive(
		StemPackRequest{
			SourceStemPath: request.SourceSproutPath,
			TargetPath:     request.TargetPath,
			Format:         request.Format,
		},
	)
}

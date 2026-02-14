package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"
)

func sproutsWipe_inner(ctx *task.ExecutionContext, sprouts *SafeSproutsReference) (error, any) {
	sproutsInfo, err := os.Lstat(sprouts.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return err, nil
	}

	sproutsMode := sproutsInfo.Mode()
	if sproutsMode|0o200 != sproutsMode {
		err, _ = fs2.Chmod(
			ctx,
			fs2.ChmodRequest{
				Path:     sprouts.BasePath,
				Mode:     sproutsMode | 0o200,
				SkipLock: true,
			},
		)
		if err != nil {
			return err, nil
		}
		defer fs2.Chmod(
			ctx,
			fs2.ChmodRequest{
				Path:     sprouts.BasePath,
				Mode:     sproutsMode,
				SkipLock: true,
			},
		)
	}

	children, err := os.ReadDir(sprouts.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return err, nil
	}

	for _, child := range children {
		childPath := filepath.Join(sprouts.BasePath, child.Name())
		err, _ := fs2.RemoveAll(ctx, childPath)
		if err != nil {
			return err, nil
		}
	}

	return nil, nil
}

func sproutsWipe(ctx *task.ExecutionContext, sprouts *SafeSproutsReference) error {
	wipeTask := fs2.WithFileLock(
		sproutsWipe_inner,
		func(ctx *task.ExecutionContext, sprouts *SafeSproutsReference) (error, string) {
			return nil, sprouts.BasePath
		},
	)

	err, _ := wipeTask(ctx, sprouts)
	return err
}

func (sprouts *SafeSproutsReference) Wipe(
	ctx *task.ExecutionContext,
) error {
	err := sproutsWipe(
		ctx,
		sprouts,
	)
	return err
}

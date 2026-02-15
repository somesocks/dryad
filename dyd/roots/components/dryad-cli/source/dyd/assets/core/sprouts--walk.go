package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var sproutsWalk_ShouldWalk = func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
	isSymlink := node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
	isDir := node.Info.IsDir()
	return nil, isDir && !isSymlink
}

var sproutsWalk_ShouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
	typePath := filepath.Join(node.Path, "dyd", "type")
	typeRaw, err := ioutil.ReadFile(typePath)
	if err == nil {
		sentinel := strings.TrimSpace(string(typeRaw))
		if sentinel == SentinelSprout.String() || sentinel == SentinelStem.String() {
			return nil, true
		}
	}

	// Legacy fallback: old sprouts were direct links to stems.
	fingerprintPath := filepath.Join(node.Path, "dyd", "fingerprint")
	_, fingerprintErr := os.Stat(fingerprintPath)
	isSprout := fingerprintErr == nil

	return nil, isSprout
}

type sproutsWalkRequest struct {
	Sprouts  *SafeSproutsReference
	OnSprout func(*task.ExecutionContext, *SafeSproutReference) (error, any)
}

func sproutsWalk(ctx *task.ExecutionContext, req sproutsWalkRequest) (error, any) {

	var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
		var unsafeRef = UnsafeSproutReference{
			BasePath: node.Path,
			Sprouts:  req.Sprouts,
		}
		var safeRef *SafeSproutReference
		var err error

		err, safeRef = unsafeRef.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, _ = req.OnSprout(ctx, safeRef)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}
	onMatch = fs2.ConditionalWalkAction(
		onMatch,
		sproutsWalk_ShouldMatch,
	)

	var err error

	err, _ = fs2.Walk6(
		ctx,
		fs2.Walk6Request{
			BasePath:   req.Sprouts.BasePath,
			Path:       req.Sprouts.BasePath,
			VPath:      req.Sprouts.BasePath,
			ShouldWalk: sproutsWalk_ShouldWalk,
			OnPreMatch: onMatch,
		},
	)
	if err != nil {
		return err, nil
	}

	return nil, nil
}

type SproutsWalkRequest struct {
	OnSprout func(*task.ExecutionContext, *SafeSproutReference) (error, any)
}

func (sprouts *SafeSproutsReference) Walk(
	ctx *task.ExecutionContext,
	req SproutsWalkRequest,
) error {
	err, _ := sproutsWalk(
		ctx,
		sproutsWalkRequest{
			Sprouts:  sprouts,
			OnSprout: req.OnSprout,
		},
	)
	return err
}

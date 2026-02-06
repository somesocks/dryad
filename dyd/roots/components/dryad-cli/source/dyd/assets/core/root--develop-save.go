package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	zlog "github.com/rs/zerolog/log"
)

type rootDevelopFileState struct {
	Kind       string
	Hash       string
	Mode       fs.FileMode
	LinkTarget string
}

type rootDevelopStatusEntry struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type rootDevelopCollectSpec struct {
	BasePath    string
	RelPrefix   string
	ApplyIgnore bool
}

func rootDevelop_collectState(
	ctx *task.ExecutionContext,
	spec rootDevelopCollectSpec,
) (map[string]rootDevelopFileState, error) {
	var err error
	var states = make(map[string]rootDevelopFileState)
	var statesMutex sync.Mutex

	if ctx == nil {
		ctx = task.DEFAULT_CONTEXT
	}
	ctx = &task.ExecutionContext{
		ConcurrencyChannel: ctx.ConcurrencyChannel,
	}

	basePath, err := filepath.Abs(spec.BasePath)
	if err != nil {
		return nil, err
	}

	shouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		if node.Info == nil {
			return nil, false
		}
		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil, false
		}
		if !node.Info.IsDir() {
			return nil, false
		}

		if spec.ApplyIgnore {
			parentDir := filepath.Dir(node.VPath)
			err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
				BasePath: basePath,
				Path:     parentDir,
			})
			if err != nil {
				return err, false
			}

			err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, true))
			if err != nil {
				return err, false
			}
			if match {
				return nil, false
			}
		}

		return nil, true
	}

	onMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		relPath, err := filepath.Rel(basePath, node.VPath)
		if err != nil {
			return err, nil
		}
		if relPath == "." {
			return nil, nil
		}

		if spec.ApplyIgnore {
			parentDir := filepath.Dir(node.VPath)
			err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
				BasePath: basePath,
				Path:     parentDir,
			})
			if err != nil {
				return err, nil
			}

			err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, node.Info.IsDir()))
			if err != nil {
				return err, nil
			}
			if match {
				return nil, nil
			}
		}

		key := filepath.Join(spec.RelPrefix, relPath)
		mode := node.Info.Mode()

		switch {
		case mode.IsDir():
			return nil, nil
		case mode&os.ModeSymlink == os.ModeSymlink:
			_, hash, err := linkHash(node.Path)
			if err != nil {
				return err, nil
			}
			linkTarget, err := os.Readlink(node.Path)
			if err != nil {
				return err, nil
			}
			statesMutex.Lock()
			states[key] = rootDevelopFileState{
				Kind:       "symlink",
				Hash:       hash,
				Mode:       mode,
				LinkTarget: linkTarget,
			}
			statesMutex.Unlock()
		case mode.IsRegular():
			_, hash, err := fileHash(node.Path)
			if err != nil {
				return err, nil
			}
			statesMutex.Lock()
			states[key] = rootDevelopFileState{
				Kind: "file",
				Hash: hash,
				Mode: mode,
			}
			statesMutex.Unlock()
		default:
			return fmt.Errorf("rootDevelop_collectState: unsupported file type: %s", node.Path), nil
		}

		return nil, nil
	}

	err, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:   basePath,
			Path:       basePath,
			VPath:      basePath,
			ShouldWalk: shouldWalk,
			OnPreMatch: onMatch,
		},
	)
	if err != nil {
		return nil, err
	}

	return states, nil
}

func rootDevelop_collectAll(
	ctx *task.ExecutionContext,
	rootPath string,
) (map[string]rootDevelopFileState, error) {
	specs := []rootDevelopCollectSpec{
		{
			BasePath:    filepath.Join(rootPath, "dyd", "assets"),
			RelPrefix:   filepath.Join("dyd", "assets"),
			ApplyIgnore: true,
		},
		{
			BasePath:  filepath.Join(rootPath, "dyd", "commands"),
			RelPrefix: filepath.Join("dyd", "commands"),
		},
		{
			BasePath:  filepath.Join(rootPath, "dyd", "docs"),
			RelPrefix: filepath.Join("dyd", "docs"),
		},
		{
			BasePath:  filepath.Join(rootPath, "dyd", "traits"),
			RelPrefix: filepath.Join("dyd", "traits"),
		},
		{
			BasePath:  filepath.Join(rootPath, "dyd", "secrets"),
			RelPrefix: filepath.Join("dyd", "secrets"),
		},
		{
			BasePath:  filepath.Join(rootPath, "dyd", "requirements"),
			RelPrefix: filepath.Join("dyd", "requirements"),
		},
	}

	all := make(map[string]rootDevelopFileState)
	for _, spec := range specs {
		exists, err := fileExists(spec.BasePath)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}

		states, err := rootDevelop_collectState(ctx, spec)
		if err != nil {
			return nil, err
		}
		for k, v := range states {
			all[k] = v
		}
	}

	return all, nil
}

func rootDevelop_stateEqual(a *rootDevelopFileState, b *rootDevelopFileState) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Kind != b.Kind {
		return false
	}
	if a.Kind == "dir" {
		return true
	}
	return a.Hash == b.Hash
}

func rootDevelop_applyFile(
	srcPath string,
	destPath string,
	state rootDevelopFileState,
) error {
	parent := filepath.Dir(destPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}

	_ = os.Remove(destPath)

	switch state.Kind {
	case "file":
		return rootDevelop_copyFile(srcPath, destPath, state.Mode)
	case "symlink":
		return os.Symlink(state.LinkTarget, destPath)
	case "dir":
		return os.MkdirAll(destPath, state.Mode)
	default:
		return fmt.Errorf("rootDevelop_applyFile: unsupported kind %s", state.Kind)
	}
}

func rootDevelop_saveChanges(
	ctx *task.ExecutionContext,
	rootPath string,
	workspacePath string,
	snapshotStemPath string,
) ([]string, error) {
	rootStates, err := rootDevelop_collectAll(ctx, rootPath)
	if err != nil {
		return nil, err
	}
	workspaceStates, err := rootDevelop_collectAll(ctx, workspacePath)
	if err != nil {
		return nil, err
	}
	snapshot, err := rootDevelop_collectAll(ctx, snapshotStemPath)
	if err != nil {
		return nil, err
	}

	keys := map[string]struct{}{}
	for k := range snapshot {
		keys[k] = struct{}{}
	}
	for k := range rootStates {
		keys[k] = struct{}{}
	}
	for k := range workspaceStates {
		keys[k] = struct{}{}
	}

	conflicts := []string{}
	changed := []string{}

	for key := range keys {
		var sPtr, rPtr, wPtr *rootDevelopFileState
		if s, ok := snapshot[key]; ok {
			sPtr = &s
		}
		if r, ok := rootStates[key]; ok {
			rPtr = &r
		}
		if w, ok := workspaceStates[key]; ok {
			wPtr = &w
		}

		workspaceChanged := !rootDevelop_stateEqual(wPtr, sPtr)
		rootChanged := !rootDevelop_stateEqual(rPtr, sPtr)

		if !workspaceChanged {
			continue
		}

		changed = append(changed, key)

		if !rootChanged || rootDevelop_stateEqual(wPtr, rPtr) {
			destPath := filepath.Join(rootPath, key)
			srcPath := filepath.Join(workspacePath, key)

			if wPtr == nil {
				if rPtr != nil && rPtr.Kind == "dir" {
					continue
				}
				err := os.Remove(destPath)
				if err != nil && !os.IsNotExist(err) {
					return conflicts, err
				}
				continue
			}

			err := rootDevelop_applyFile(srcPath, destPath, *wPtr)
			if err != nil {
				return conflicts, err
			}
			continue
		}

		conflicts = append(conflicts, key)
	}

	for _, path := range conflicts {
		zlog.Warn().Str("path", path).Msg("root develop save conflict")
	}

	return conflicts, nil
}

func rootDevelop_statusChanges(
	ctx *task.ExecutionContext,
	rootPath string,
	workspacePath string,
	snapshotStemPath string,
) ([]rootDevelopStatusEntry, error) {
	rootStates, err := rootDevelop_collectAll(ctx, rootPath)
	if err != nil {
		return nil, err
	}
	workspaceStates, err := rootDevelop_collectAll(ctx, workspacePath)
	if err != nil {
		return nil, err
	}
	snapshot, err := rootDevelop_collectAll(ctx, snapshotStemPath)
	if err != nil {
		return nil, err
	}

	keys := map[string]struct{}{}
	for k := range snapshot {
		keys[k] = struct{}{}
	}
	for k := range rootStates {
		keys[k] = struct{}{}
	}
	for k := range workspaceStates {
		keys[k] = struct{}{}
	}

	entries := []rootDevelopStatusEntry{}

	for key := range keys {
		var sPtr, rPtr, wPtr *rootDevelopFileState
		if s, ok := snapshot[key]; ok {
			sPtr = &s
		}
		if r, ok := rootStates[key]; ok {
			rPtr = &r
		}
		if w, ok := workspaceStates[key]; ok {
			wPtr = &w
		}

		workspaceChanged := !rootDevelop_stateEqual(wPtr, sPtr)
		rootChanged := !rootDevelop_stateEqual(rPtr, sPtr)

		if !workspaceChanged {
			continue
		}

		conflict := rootChanged && !rootDevelop_stateEqual(wPtr, rPtr)
		entry := rootDevelopStatusEntry{
			Code: " M",
			Path: key,
		}
		if conflict {
			entry.Code = "UU"
		} else if sPtr == nil && wPtr != nil {
			entry.Code = "??"
		} else if sPtr != nil && wPtr == nil {
			entry.Code = " D"
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

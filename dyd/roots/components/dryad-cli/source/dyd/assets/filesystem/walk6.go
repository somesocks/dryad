package fs2

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"dryad/task"
)

type Walk6Node struct {
	BasePath string
	Path     string
	VPath    string
	Info     fs.FileInfo
}

type WalkDecision func (ctx *task.ExecutionContext, node Walk6Node) (error, bool)

type WalkAction func (ctx *task.ExecutionContext, node Walk6Node) (error, any)

func ConditionalWalkAction(
	action WalkAction,
	decision WalkDecision,
) WalkAction {
	return func (ctx *task.ExecutionContext, node Walk6Node) (error, any) {
		err, match := decision(ctx, node)
		if err != nil {
			return err, nil
		} else if !match {
			return nil, nil
		} else {
			err, _ = action(ctx, node)
			return err, nil
		}
	}
}

type Walk6Request struct {
	BasePath    string
	Path        string
	VPath       string
	ShouldWalk WalkDecision
	OnPreMatch WalkAction
	OnPostMatch WalkAction
}

var defaultShouldWalk6 WalkDecision = func (ctx *task.ExecutionContext, node Walk6Node) (error, bool) {
	return nil, true
}

var defaultOnPreMatch6 WalkAction = func (ctx *task.ExecutionContext, node Walk6Node) (error, any) {
	return nil, nil
}

var defaultOnPostMatch6 WalkAction = func (ctx *task.ExecutionContext, node Walk6Node) (error, any) {
	return nil, nil
}


func _walk6(ctx *task.ExecutionContext, request Walk6Request) (error) {

	var err error
	var info fs.FileInfo

	// read file info from real path,
	// without resolving symlink
	info, err = os.Lstat(request.Path)
	if err != nil {
		return err
	}

	// --------------------------------------
	// STEP 1
	// run pre-matching (breadth-first) 
	err, _ = request.OnPreMatch(
		ctx,
		Walk6Node{
			BasePath: request.BasePath,
			Path:     request.Path,
			VPath:    request.VPath,
			Info:     info,
		},	
	)
	if err != nil {
		return err
	}

	// --------------------------------------
	// STEP 2
	// walk based on type

	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		// STEP 2.1 - symlink walking

		// check if we should walk through the link
		err, shouldWalk := request.ShouldWalk(
			ctx,
			Walk6Node{
				BasePath: request.BasePath,
				Path:     request.Path,
				VPath:    request.VPath,
				Info:     info,
			},
		)
		if err != nil {
			return err
		}

		// if we should, resolve the link to it's real path
		if shouldWalk {
			linkPath, err := os.Readlink(request.Path)
			if err != nil {
				return err
			}

			// clean up relative links
			if !filepath.IsAbs(linkPath) {
				linkPath = filepath.Clean(filepath.Join(filepath.Dir(request.Path), linkPath))
			}

			// walk through to the real link
			// update the real path, but not the virtual path
			err = _walk6(ctx, Walk6Request{
				BasePath:    request.BasePath,
				Path:        linkPath,
				VPath:       request.VPath,
				ShouldWalk: request.ShouldWalk,
				OnPreMatch:     request.OnPreMatch,
				OnPostMatch:     request.OnPostMatch,
			})
			if err != nil {
				return err
			}
		}

	} else if info.IsDir() {
		// STEP 2.2 - directory walking

		// check if we should walk through the dir
		err, shouldWalk := request.ShouldWalk(
			ctx,
			Walk6Node{
				BasePath: request.BasePath,
				Path:     request.Path,
				VPath:    request.VPath,
				Info:     info,
			},
		)
		if err != nil {
			return err
		}

		if shouldWalk {
			dir, err := os.Open(request.Path)
			if err != nil && err != io.EOF {
				return err
			}
			defer dir.Close()

			var entries []fs.DirEntry
			entries, err = dir.ReadDir(128)
			for err != io.EOF {
				if err != nil {
					return err
				}

				parallelWalk := func (ctx *task.ExecutionContext, entry fs.DirEntry) (error, any) {
					err := _walk6(
						ctx,
						Walk6Request{
							BasePath:    request.BasePath,
							Path:        filepath.Join(request.Path, entry.Name()),
							VPath:       filepath.Join(request.VPath, entry.Name()),
							ShouldWalk:  request.ShouldWalk,
							OnPreMatch:  request.OnPreMatch,
							OnPostMatch:  request.OnPostMatch,
						},
					)
					return err, nil
				}

				err, _ = task.ParallelMap(parallelWalk)(ctx, entries)
				if err != nil {
					return err
				}
				entries, err = dir.ReadDir(128)
			}
		}		
	}


	// --------------------------------------
	// STEP 3
	// run post-matching (depth-first) 
	err, _ = request.OnPostMatch(
		ctx,
		Walk6Node{
			BasePath: request.BasePath,
			Path:     request.Path,
			VPath:    request.VPath,
			Info:     info,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func Walk6(ctx *task.ExecutionContext, request Walk6Request) (error, any) {

	if request.ShouldWalk == nil {
		request.ShouldWalk = defaultShouldWalk6
	}
	
	if request.OnPreMatch == nil {
		request.OnPreMatch = defaultOnPreMatch6
	}

	if  request.OnPostMatch == nil {
		request.OnPostMatch = defaultOnPostMatch6
	}

	err := _walk6(ctx, request)
	return err, nil
}

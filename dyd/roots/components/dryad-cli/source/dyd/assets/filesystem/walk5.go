package fs2

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"dryad/task"
)

var defaultShouldCrawl3 = func(node Walk5Node) (bool, error) {
	return true, nil
}

var defaultShouldMatch3 = func(node Walk5Node) (bool, error) {
	return true, nil
}

var defaultOnMatch3 = func(ctx *task.ExecutionContext, node Walk5Node) (error, any) {
	return nil, nil
}

var defaultOnError3 = func(err error, node Walk5Node) error {
	return err
}

type Walk5Request struct {
	Path        string
	VPath       string
	BasePath    string
	ShouldCrawl func(node Walk5Node) (bool, error)
	ShouldMatch func(node Walk5Node) (bool, error)
	OnMatch     func(ctx *task.ExecutionContext, node Walk5Node) (error, any)
	OnError     func(err error, node Walk5Node) error
}

type Walk5Node struct {
	Path     string
	VPath    string
	BasePath string
	Info     fs.FileInfo
}


func _dfsWalk3(ctx *task.ExecutionContext, request Walk5Request) error {
	var err error
	var info fs.FileInfo

	// read file info from real path,
	// without resolving symlink
	info, err = os.Lstat(request.Path)
	if err != nil {
		err = request.OnError(err, Walk5Node{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			return err
		}
	}

	// info could be nil here because of error swallowing in OnError
	if info != nil {
		// if the file is a symlink,
		// check if we should crawl through to the real node
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {

			// check if we should crawl through the link
			shouldCrawl, err := request.ShouldCrawl(Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk5Node{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil {
					return err
				}
			}

			// if we should, resolve the link to it's real path
			if shouldCrawl {
				linkPath, err := os.Readlink(request.Path)
				if err != nil {
					err = request.OnError(err, Walk5Node{
						Path:     request.Path,
						VPath:    request.VPath,
						BasePath: request.BasePath,
						Info:     info,
					})
					if err != nil {
						return err
					}
				}
				// clean up relative links
				if !filepath.IsAbs(linkPath) {
					linkPath = filepath.Clean(filepath.Join(filepath.Dir(request.Path), linkPath))
				}

				// crawl through to the real link
				// update the real path, but not the virtual path
				err = _dfsWalk3(ctx, Walk5Request{
					Path:        linkPath,
					VPath:       request.VPath,
					BasePath:    request.BasePath,
					ShouldCrawl: request.ShouldCrawl,
					ShouldMatch: request.ShouldMatch,
					OnMatch:     request.OnMatch,
					OnError:     request.OnError,
				})
				if err != nil {
					err = request.OnError(err, Walk5Node{
						Path:     request.Path,
						VPath:    request.VPath,
						BasePath: request.BasePath,
						Info:     info,
					})
					if err != nil {
						return err
					}
				}
			}
		} else if info.IsDir() {

			// check if we should crawl through the dir
			shouldCrawl, err := request.ShouldCrawl(Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk5Node{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil {
					return err
				}
			}

			if shouldCrawl {
				dir, err := os.Open(request.Path)
				err = request.OnError(err, Walk5Node{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil && err != io.EOF {
					err = request.OnError(err, Walk5Node{
						Path:     request.Path,
						VPath:    request.VPath,
						BasePath: request.BasePath,
						Info:     info,
					})
					if err != nil {
						return err
					}
				}
				defer dir.Close()

				var entries []fs.DirEntry
				entries, err = dir.ReadDir(100)
				for err != io.EOF {
					if err != nil {
						err = request.OnError(err, Walk5Node{
							Path:     request.Path,
							VPath:    request.VPath,
							BasePath: request.BasePath,
							Info:     info,
						})
						if err != nil {
							return err
						}
					}
					err, _ = task.ParallelMap(
						func (ctx *task.ExecutionContext, entry fs.DirEntry) (error, any) {
							var err error
							err = _dfsWalk3(
								ctx,
								Walk5Request{
									Path:        filepath.Join(request.Path, entry.Name()),
									VPath:       filepath.Join(request.VPath, entry.Name()),
									BasePath:    request.BasePath,
									ShouldCrawl: request.ShouldCrawl,
									ShouldMatch: request.ShouldMatch,
									OnMatch:     request.OnMatch,
									OnError:     request.OnError,
								},
							)
							if err != nil {
								err = request.OnError(err, Walk5Node{
									Path:     request.Path,
									VPath:    request.VPath,
									BasePath: request.BasePath,
									Info:     info,
								})
								if err != nil {
									return err, nil
								}
							}
							return nil, nil
						},
					)(ctx, entries)
					if err != nil {
						return err
					}
					entries, err = dir.ReadDir(100)
				}
			}
		}
	}

	shouldMatch, err := request.ShouldMatch(Walk5Node{
		Path:     request.Path,
		VPath:    request.VPath,
		BasePath: request.BasePath,
		Info:     info,
	})
	if err != nil {
		err = request.OnError(err, Walk5Node{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			return err
		}
	}

	if shouldMatch {
		err, _ = request.OnMatch(
			ctx,
			Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			},
		)
		if err != nil {
			err = request.OnError(err, Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DFSWalk3(ctx *task.ExecutionContext, request Walk5Request) error {

	if request.ShouldCrawl == nil {
		request.ShouldCrawl = defaultShouldCrawl3
	}

	if request.ShouldMatch == nil {
		request.ShouldMatch = defaultShouldMatch3
	}

	if request.OnMatch == nil {
		request.OnMatch = defaultOnMatch3
	}

	if request.OnError == nil {
		request.OnError = defaultOnError3
	}

	return _dfsWalk3(ctx, request)
}

func _bfsWalk3(ctx *task.ExecutionContext, request Walk5Request) error {

	var err error
	var info fs.FileInfo

	// read file info from real path,
	// without resolving symlink
	info, err = os.Lstat(request.Path)
	if err != nil {
		err = request.OnError(err, Walk5Node{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			return err
		}
	}

	shouldMatch, err := request.ShouldMatch(Walk5Node{
		Path:     request.Path,
		VPath:    request.VPath,
		BasePath: request.BasePath,
		Info:     info,
	})
	if err != nil {
		err = request.OnError(err, Walk5Node{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			return err
		}
	}

	if shouldMatch {
		err, _ = request.OnMatch(
			ctx,
			Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			},
		)
		if err != nil {
			err = request.OnError(err, Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				return err
			}
		}
	}

	// info could be nil here because of error swallowing in OnError
	if info != nil {
		// if the file is a symlink,
		// check if we should crawl through to the real node
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {

			// check if we should crawl through the link
			shouldCrawl, err := request.ShouldCrawl(Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk5Node{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil {
					return err
				}
			}

			// if we should, resolve the link to it's real path
			if shouldCrawl {
				linkPath, err := os.Readlink(request.Path)
				if err != nil {
					err = request.OnError(err, Walk5Node{
						Path:     request.Path,
						VPath:    request.VPath,
						BasePath: request.BasePath,
						Info:     info,
					})
					if err != nil {
						return err
					}
				}
				// clean up relative links
				if !filepath.IsAbs(linkPath) {
					linkPath = filepath.Clean(filepath.Join(filepath.Dir(request.Path), linkPath))
				}

				// crawl through to the real link
				// update the real path, but not the virtual path
				err = _bfsWalk3(
					ctx,
					Walk5Request{
						Path:        linkPath,
						VPath:       request.VPath,
						BasePath:    request.BasePath,
						ShouldCrawl: request.ShouldCrawl,
						ShouldMatch: request.ShouldMatch,
						OnMatch:     request.OnMatch,
						OnError:     request.OnError,
					},
				)
				if err != nil {
					err = request.OnError(err, Walk5Node{
						Path:     request.Path,
						VPath:    request.VPath,
						BasePath: request.BasePath,
						Info:     info,
					})
					if err != nil {
						return err
					}
				}
			}
		} else if info.IsDir() {

			// check if we should crawl through the dir
			shouldCrawl, err := request.ShouldCrawl(Walk5Node{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk5Node{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil {
					return err
				}
			}

			if shouldCrawl {
				dir, err := os.Open(request.Path)
				err = request.OnError(err, Walk5Node{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil && err != io.EOF {
					err = request.OnError(err, Walk5Node{
						Path:     request.Path,
						VPath:    request.VPath,
						BasePath: request.BasePath,
						Info:     info,
					})
					if err != nil {
						return err
					}
				}
				defer dir.Close()

				var entries []fs.DirEntry
				entries, err = dir.ReadDir(100)
				for err != io.EOF {
					if err != nil {
						err = request.OnError(err, Walk5Node{
							Path:     request.Path,
							VPath:    request.VPath,
							BasePath: request.BasePath,
							Info:     info,
						})
						if err != nil {
							return err
						}
					}
					err, _ = task.ParallelMap(
						func (ctx *task.ExecutionContext, entry fs.DirEntry) (error, any) {
							var err error
							err = _bfsWalk3(
								ctx,
								Walk5Request{
									Path:        filepath.Join(request.Path, entry.Name()),
									VPath:       filepath.Join(request.VPath, entry.Name()),
									BasePath:    request.BasePath,
									ShouldCrawl: request.ShouldCrawl,
									ShouldMatch: request.ShouldMatch,
									OnMatch:     request.OnMatch,
									OnError:     request.OnError,
								},
							)
							if err != nil {
								err = request.OnError(err, Walk5Node{
									Path:     request.Path,
									VPath:    request.VPath,
									BasePath: request.BasePath,
									Info:     info,
								})
								if err != nil {
									return err, nil
								}
							}
							return nil, nil
						},
					)(ctx, entries)
					if err != nil {
						return err
					}
					entries, err = dir.ReadDir(100)
				}
			}
		} else {
		}
	}

	return nil
}

func BFSWalk3(ctx *task.ExecutionContext, request Walk5Request) (error, any) {

	if request.ShouldCrawl == nil {
		request.ShouldCrawl = defaultShouldCrawl3
	}

	if request.ShouldMatch == nil {
		request.ShouldMatch = defaultShouldMatch3
	}

	if request.OnMatch == nil {
		request.OnMatch = defaultOnMatch3
	}

	if request.OnError == nil {
		request.OnError = defaultOnError3
	}

	return _bfsWalk3(ctx, request), nil
}

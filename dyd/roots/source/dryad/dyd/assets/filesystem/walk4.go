package fs2

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var defaultShouldCrawl2 = func(context Walk4Context) (bool, error) {
	return true, nil
}

var defaultShouldMatch2 = func(context Walk4Context) (bool, error) {
	return true, nil
}

var defaultOnMatch2 = func(context Walk4Context) error {
	return nil
}

var defaultOnError2 = func(err error, context Walk4Context) error {
	return err
}

type Walk4Request struct {
	Path        string
	VPath       string
	BasePath    string
	ShouldCrawl func(context Walk4Context) (bool, error)
	ShouldMatch func(context Walk4Context) (bool, error)
	OnMatch     func(context Walk4Context) error
	OnError     func(err error, context Walk4Context) error
}

type Walk4Context struct {
	Path     string
	VPath    string
	BasePath string
	Info     fs.FileInfo
}

func _dfsWalk2(request Walk4Request) error {
	var err error
	var info fs.FileInfo

	// read file info from real path,
	// without resolving symlink
	info, err = os.Lstat(request.Path)
	if err != nil {
		err = request.OnError(err, Walk4Context{
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
			shouldCrawl, err := request.ShouldCrawl(Walk4Context{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk4Context{
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
					err = request.OnError(err, Walk4Context{
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
				err = _dfsWalk2(Walk4Request{
					Path:        linkPath,
					VPath:       request.VPath,
					BasePath:    request.BasePath,
					ShouldCrawl: request.ShouldCrawl,
					ShouldMatch: request.ShouldMatch,
					OnMatch:     request.OnMatch,
					OnError:     request.OnError,
				})
				if err != nil {
					err = request.OnError(err, Walk4Context{
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
			shouldCrawl, err := request.ShouldCrawl(Walk4Context{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk4Context{
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
				err = request.OnError(err, Walk4Context{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil && err != io.EOF {
					err = request.OnError(err, Walk4Context{
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
						err = request.OnError(err, Walk4Context{
							Path:     request.Path,
							VPath:    request.VPath,
							BasePath: request.BasePath,
							Info:     info,
						})
						if err != nil {
							return err
						}
					}
					for _, entry := range entries {
						err = _dfsWalk2(Walk4Request{
							Path:        filepath.Join(request.Path, entry.Name()),
							VPath:       filepath.Join(request.VPath, entry.Name()),
							BasePath:    request.BasePath,
							ShouldCrawl: request.ShouldCrawl,
							ShouldMatch: request.ShouldMatch,
							OnMatch:     request.OnMatch,
							OnError:     request.OnError,
						})
						if err != nil {
							err = request.OnError(err, Walk4Context{
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
					entries, err = dir.ReadDir(100)
				}
			}
		}
	}

	shouldMatch, err := request.ShouldMatch(Walk4Context{
		Path:     request.Path,
		VPath:    request.VPath,
		BasePath: request.BasePath,
		Info:     info,
	})
	if err != nil {
		err = request.OnError(err, Walk4Context{
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
		err = request.OnMatch(Walk4Context{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			err = request.OnError(err, Walk4Context{
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

func DFSWalk2(request Walk4Request) error {

	if request.ShouldCrawl == nil {
		request.ShouldCrawl = defaultShouldCrawl2
	}

	if request.ShouldMatch == nil {
		request.ShouldMatch = defaultShouldMatch2
	}

	if request.OnMatch == nil {
		request.OnMatch = defaultOnMatch2
	}

	if request.OnError == nil {
		request.OnError = defaultOnError2
	}

	return _dfsWalk2(request)
}

func _bfsWalk2(request Walk4Request) error {

	var err error
	var info fs.FileInfo

	// read file info from real path,
	// without resolving symlink
	info, err = os.Lstat(request.Path)
	if err != nil {
		err = request.OnError(err, Walk4Context{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			return err
		}
	}

	shouldMatch, err := request.ShouldMatch(Walk4Context{
		Path:     request.Path,
		VPath:    request.VPath,
		BasePath: request.BasePath,
		Info:     info,
	})
	if err != nil {
		err = request.OnError(err, Walk4Context{
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
		err = request.OnMatch(Walk4Context{
			Path:     request.Path,
			VPath:    request.VPath,
			BasePath: request.BasePath,
			Info:     info,
		})
		if err != nil {
			err = request.OnError(err, Walk4Context{
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
			shouldCrawl, err := request.ShouldCrawl(Walk4Context{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk4Context{
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
					err = request.OnError(err, Walk4Context{
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
				err = _bfsWalk2(Walk4Request{
					Path:        linkPath,
					VPath:       request.VPath,
					BasePath:    request.BasePath,
					ShouldCrawl: request.ShouldCrawl,
					ShouldMatch: request.ShouldMatch,
					OnMatch:     request.OnMatch,
					OnError:     request.OnError,
				})
				if err != nil {
					err = request.OnError(err, Walk4Context{
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
			shouldCrawl, err := request.ShouldCrawl(Walk4Context{
				Path:     request.Path,
				VPath:    request.VPath,
				BasePath: request.BasePath,
				Info:     info,
			})
			if err != nil {
				err = request.OnError(err, Walk4Context{
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
				err = request.OnError(err, Walk4Context{
					Path:     request.Path,
					VPath:    request.VPath,
					BasePath: request.BasePath,
					Info:     info,
				})
				if err != nil && err != io.EOF {
					err = request.OnError(err, Walk4Context{
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
						err = request.OnError(err, Walk4Context{
							Path:     request.Path,
							VPath:    request.VPath,
							BasePath: request.BasePath,
							Info:     info,
						})
						if err != nil {
							return err
						}
					}
					for _, entry := range entries {
						err = _bfsWalk2(Walk4Request{
							Path:        filepath.Join(request.Path, entry.Name()),
							VPath:       filepath.Join(request.VPath, entry.Name()),
							BasePath:    request.BasePath,
							ShouldCrawl: request.ShouldCrawl,
							ShouldMatch: request.ShouldMatch,
							OnMatch:     request.OnMatch,
							OnError:     request.OnError,
						})
						if err != nil {
							err = request.OnError(err, Walk4Context{
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
					entries, err = dir.ReadDir(100)
				}
			}
		} else {
		}
	}

	return nil
}

func BFSWalk2(request Walk4Request) error {

	if request.ShouldCrawl == nil {
		request.ShouldCrawl = defaultShouldCrawl2
	}

	if request.ShouldMatch == nil {
		request.ShouldMatch = defaultShouldMatch2
	}

	if request.OnMatch == nil {
		request.OnMatch = defaultOnMatch2
	}

	if request.OnError == nil {
		request.OnError = defaultOnError2
	}

	return _bfsWalk2(request)
}

package fs2

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var defaultShouldCrawl = func(path string, info fs.FileInfo) (bool, error) {
	return true, nil
}

var defaultShouldMatch = func(path string, info fs.FileInfo) (bool, error) {
	return true, nil
}

var defaultOnMatch = func(path string, info fs.FileInfo) error {
	return nil
}

var defaultOnError = func(err error, path string, info fs.FileInfo) error {
	return err
}

type Walk3Request struct {
	BasePath    string
	ShouldCrawl func(path string, info fs.FileInfo) (bool, error)
	ShouldMatch func(path string, info fs.FileInfo) (bool, error)
	OnMatch     func(path string, info fs.FileInfo) error
	OnError     func(err error, path string, info fs.FileInfo) error
}

func _dfsWalk(context Walk3Request, path string) error {
	var err error
	var info fs.FileInfo

	info, err = os.Lstat(path)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}

	// info could still be nil here because of error swallowing in OnError
	if info != nil && info.Mode()&os.ModeSymlink == os.ModeSymlink {
		shouldCrawl, err := context.ShouldCrawl(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}

		if shouldCrawl {
			var linkPath string
			linkPath, err = os.Readlink(path)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
			if !filepath.IsAbs(linkPath) {
				linkPath = filepath.Join(
					filepath.Dir(path),
					linkPath,
				)
			}

			err = _dfsWalk(context, linkPath)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
		}

	} else if info != nil && info.IsDir() {
		shouldCrawl, err := context.ShouldCrawl(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}

		if shouldCrawl {
			var dir *os.File
			dir, err = os.Open(path)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
			defer dir.Close()

			var entries []fs.DirEntry

			entries, err = dir.ReadDir(100)
			if err != nil && err != io.EOF {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}

			for len(entries) > 0 {
				for _, entry := range entries {
					err = _dfsWalk(context, filepath.Join(path, entry.Name()))
					if err != nil {
						err = context.OnError(err, path, info)
						if err != nil {
							return err
						}
					}
				}

				entries, err = dir.ReadDir(100)
				if err != nil && err != io.EOF {
					err = context.OnError(err, path, info)
					if err != nil {
						return err
					}
				}
			}
		}

	}

	shouldMatch, err := context.ShouldMatch(path, info)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}

	if shouldMatch {
		err = context.OnMatch(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DFSWalk(request Walk3Request) error {
	if request.ShouldCrawl == nil {
		request.ShouldCrawl = defaultShouldCrawl
	}

	if request.ShouldMatch == nil {
		request.ShouldMatch = defaultShouldMatch
	}

	if request.OnMatch == nil {
		request.OnMatch = defaultOnMatch
	}

	if request.OnError == nil {
		request.OnError = defaultOnError
	}

	return _dfsWalk(request, request.BasePath)
}

func _bfsWalk(context Walk3Request, path string) error {
	var err error
	var info fs.FileInfo

	info, err = os.Lstat(path)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}

	shouldMatch, err := context.ShouldMatch(path, info)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}

	if shouldMatch {
		err = context.OnMatch(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}
	}

	// info could still be nil here because of error swallowing in OnError
	if info != nil && info.Mode()&os.ModeSymlink == os.ModeSymlink {
		shouldCrawl, err := context.ShouldCrawl(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}

		if shouldCrawl {
			var linkPath string
			linkPath, err = os.Readlink(path)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
			if !filepath.IsAbs(linkPath) {
				linkPath = filepath.Join(
					filepath.Dir(path),
					linkPath,
				)
			}

			err = _bfsWalk(context, linkPath)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
		}

	} else if info != nil && info.IsDir() {
		shouldCrawl, err := context.ShouldCrawl(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}

		if shouldCrawl {
			var dir *os.File
			dir, err = os.Open(path)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
			defer dir.Close()

			var entries []fs.DirEntry

			entries, err = dir.ReadDir(100)
			if err != nil && err != io.EOF {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}

			for len(entries) > 0 {
				for _, entry := range entries {
					err = _bfsWalk(context, filepath.Join(path, entry.Name()))
					if err != nil {
						err = context.OnError(err, path, info)
						if err != nil {
							return err
						}
					}
				}

				entries, err = dir.ReadDir(100)
				if err != nil && err != io.EOF {
					err = context.OnError(err, path, info)
					if err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}

func BFSWalk(request Walk3Request) error {
	if request.ShouldCrawl == nil {
		request.ShouldCrawl = defaultShouldCrawl
	}

	if request.ShouldMatch == nil {
		request.ShouldMatch = defaultShouldMatch
	}

	if request.OnMatch == nil {
		request.OnMatch = defaultOnMatch
	}

	if request.OnError == nil {
		request.OnError = defaultOnError
	}

	return _bfsWalk(request, request.BasePath)
}

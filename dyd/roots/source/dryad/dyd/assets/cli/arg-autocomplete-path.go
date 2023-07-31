package cli

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ArgAutoCompletePath(token string) []string {
	if token == "" {
		return []string{"./"}
	}

	var results = []string{}
	var base string
	var parent string

	if strings.HasSuffix(token, string(filepath.Separator)) {
		base = ""
		parent = token
	} else {
		base = filepath.Base(token)
		parent = filepath.Dir(token)
	}

	dir, err := os.Open(parent)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	var entries []fs.DirEntry
	entries, err = dir.ReadDir(100)
	for err != io.EOF {
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			var name = entry.Name()
			if strings.HasPrefix(name, base) {
				var isDir = entry.IsDir()
				if isDir {
					results = append(results, name+string(filepath.Separator))
				} else {
					results = append(results, name)
				}
			}
		}
		entries, err = dir.ReadDir(100)
	}

	return results
}

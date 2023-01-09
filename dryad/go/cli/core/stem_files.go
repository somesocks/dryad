package core

import (
	"fmt"
	"io/fs"
)

func StemFiles(path string) error {
	StemWalk(
		path,
		func(walk string, info fs.FileInfo, err error) error {
			// var rel, relErr = filepath.Rel(path, walk)
			// if relErr != nil {
			// 	return relErr
			// }
			fmt.Println(walk)
			return nil
		},
	)
	return nil
}

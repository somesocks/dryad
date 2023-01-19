package core

import (
	"fmt"
	"io/fs"
)

func StemFiles(path string) error {
	StemWalk(
		StemWalkArgs{
			BasePath: path,
			OnMatch: func(walk string, info fs.FileInfo, err error) error {
				if !info.IsDir() {
					fmt.Println(walk)
				}
				return nil
			},
		},
	)
	return nil
}

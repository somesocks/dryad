package core

import (
	"fmt"
	"io/fs"
	"regexp"
)

type StemFilesArgs struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

func StemFiles(args StemFilesArgs) error {
	StemWalk(
		StemWalkRequest{
			BasePath: args.BasePath,
			OnMatch: func(walk string, info fs.FileInfo, basePath string) error {
				if !info.IsDir() {
					fmt.Println(walk)
				}
				return nil
			},
		},
	)
	return nil
}

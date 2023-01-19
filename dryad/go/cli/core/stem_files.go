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
		StemWalkArgs{
			BasePath:  args.BasePath,
			MatchDeny: args.MatchDeny,
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

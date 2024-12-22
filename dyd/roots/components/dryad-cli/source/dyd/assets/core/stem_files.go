package core

import (
	fs2 "dryad/filesystem"
	"fmt"
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
			OnMatch: func(context fs2.Walk4Context) error {
				if !context.Info.IsDir() {
					fmt.Println(context.VPath)
				}
				return nil
			},
		},
	)
	return nil
}

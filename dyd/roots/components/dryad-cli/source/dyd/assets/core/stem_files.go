package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"fmt"
	"regexp"
)

type StemFilesArgs struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

func StemFiles(args StemFilesArgs) error {
	StemWalk(
		task.DEFAULT_CONTEXT,
		StemWalkRequest{
			BasePath: args.BasePath,
			OnMatch: func(node fs2.Walk5Node) error {
				if !node.Info.IsDir() {
					fmt.Println(node.VPath)
				}
				return nil
			},
		},
	)
	return nil
}

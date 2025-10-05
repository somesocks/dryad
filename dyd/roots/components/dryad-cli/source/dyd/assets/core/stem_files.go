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
			OnMatch: func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
				if !node.Info.IsDir() {
					fmt.Println(node.VPath)
				}
				return nil, nil
			},
		},
	)
	return nil
}

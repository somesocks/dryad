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

package core

import (
	"context"
	fs2 "dryad/filesystem"
	"dryad/internal/filepath"
	"fmt"
	"regexp"
)

type StemFilesArgs struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

func StemFiles(ctx context.Context, args StemFilesArgs) error {
	err, _ := StemWalk(
		ctx,
		StemWalkRequest{
			BasePath: args.BasePath,
			OnMatch: func(ctx context.Context, node fs2.Walk6Node) (error, any) {
				relPath, err := filepath.Rel(node.BasePath, node.VPath)
				if err != nil {
					return err, nil
				}

				if args.MatchDeny != nil && args.MatchDeny.Match([]byte(relPath)) {
					return nil, nil
				}

				if !node.Info.IsDir() {
					fmt.Println(node.VPath)
				}
				return nil, nil
			},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

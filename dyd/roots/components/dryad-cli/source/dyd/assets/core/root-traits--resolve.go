package core

import "context"

// fs2 "dryad/filesystem"

// zlog "github.com/rs/zerolog/log"

func (rootTraits *UnsafeRootTraitsReference) Resolve(ctx context.Context) (error, *SafeRootTraitsReference) {
	var rootTraitsExists bool
	var err error
	var safeRef SafeRootTraitsReference

	rootTraitsExists, err = fileExists(rootTraits.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootTraitsExists {
		return nil, nil
	}

	safeRef = SafeRootTraitsReference{
		BasePath: rootTraits.BasePath,
		Root:     rootTraits.Root,
	}

	return nil, &safeRef
}

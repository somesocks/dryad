package core

import (
	"path/filepath"
	// "dryad/task"
	// zlog "github.com/rs/zerolog/log"
)

func (sprouts *SafeHeapSproutsReference) Sprout(fingerprint string) *UnsafeHeapSproutReference {
	var heapSproutRef = UnsafeHeapSproutReference{
		BasePath: filepath.Join(sprouts.BasePath, fingerprint),
		Sprouts:  sprouts,
	}
	return &heapSproutRef
}

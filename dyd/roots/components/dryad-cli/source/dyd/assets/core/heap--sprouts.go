package core

import (
	"dryad/internal/filepath"
	// "dryad/task"
	// zlog "github.com/rs/zerolog/log"
)

func (heap *SafeHeapReference) Sprouts() *UnsafeHeapSproutsReference {
	var heapSproutsRef = UnsafeHeapSproutsReference{
		BasePath: filepath.Join(heap.BasePath, "sprouts"),
		Heap:     heap,
	}
	return &heapSproutsRef
}

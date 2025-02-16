package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (heap *SafeHeapReference) Derivations() (*UnsafeHeapDerivationsReference) {
	var heapDerivationsRef = UnsafeHeapDerivationsReference{
		BasePath: filepath.Join(heap.BasePath, "derivations"),
		Heap: heap,
	}
	return &heapDerivationsRef
}
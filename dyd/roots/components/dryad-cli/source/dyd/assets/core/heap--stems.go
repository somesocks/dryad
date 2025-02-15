package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (heap *SafeHeapReference) Stems() (*UnsafeHeapStemsReference) {
	var heapStemsRef = UnsafeHeapStemsReference{
		BasePath: filepath.Join(heap.BasePath, "stems"),
		Heap: heap,
	}
	return &heapStemsRef
}
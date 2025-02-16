package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (heap *SafeHeapReference) Files() (*UnsafeHeapFilesReference) {
	var heapFilesRef = UnsafeHeapFilesReference{
		BasePath: filepath.Join(heap.BasePath, "files"),
		Heap: heap,
	}
	return &heapFilesRef
}
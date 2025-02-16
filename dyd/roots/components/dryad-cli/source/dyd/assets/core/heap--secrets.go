package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (heap *SafeHeapReference) Secrets() (*UnsafeHeapSecretsReference) {
	var heapSecretsRef = UnsafeHeapSecretsReference{
		BasePath: filepath.Join(heap.BasePath, "secrets"),
		Heap: heap,
	}
	return &heapSecretsRef
}
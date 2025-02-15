package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (sg *SafeGardenReference) Heap() (*UnsafeHeapReference) {
	var heapRef = UnsafeHeapReference{
		BasePath: filepath.Join(sg.BasePath, "dyd", "heap"),
		Garden: sg,
	}
	return &heapRef
}
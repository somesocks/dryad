package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (stems *SafeHeapStemsReference) Stem(fingerprint string) (*UnsafeHeapStemReference) {
	var heapStemRef = UnsafeHeapStemReference{
		BasePath: filepath.Join(stems.BasePath, fingerprint),
		Stems: stems,
	}
	return &heapStemRef
}
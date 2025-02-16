package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (stems *SafeHeapDerivationsReference) Derivation(fingerprint string) (*UnsafeHeapDerivationReference) {
	var heapDerivationRef = UnsafeHeapDerivationReference{
		BasePath: filepath.Join(stems.BasePath, fingerprint),
		Derivations: stems,
	}
	return &heapDerivationRef
}
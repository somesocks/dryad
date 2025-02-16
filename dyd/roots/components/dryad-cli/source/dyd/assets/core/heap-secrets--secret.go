package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (stems *SafeHeapSecretsReference) Secrets(fingerprint string) (*UnsafeHeapSecretReference) {
	var heapSecretRef = UnsafeHeapSecretReference{
		BasePath: filepath.Join(stems.BasePath, fingerprint),
		Secrets: stems,
	}
	return &heapSecretRef
}
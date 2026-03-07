package core

import (
	"path/filepath"
	"strings"
)

func (stems *SafeHeapStemsReference) Stem(fingerprint string) *UnsafeHeapStemReference {
	fingerprint = strings.TrimSpace(fingerprint)
	encoded := strings.TrimPrefix(fingerprint, fingerprintVersionV2+"-")
	basePath := filepath.Join(stems.BasePath, fingerprintVersionV2, encoded)
	var heapStemRef = UnsafeHeapStemReference{
		BasePath:    basePath,
		Fingerprint: fingerprint,
		Stems:       stems,
	}
	return &heapStemRef
}

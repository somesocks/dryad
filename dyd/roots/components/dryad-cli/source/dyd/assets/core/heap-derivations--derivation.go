package core

import (
	"path/filepath"
	"strings"
)

func (stems *SafeHeapDerivationsReference) Derivation(fingerprint string) *UnsafeHeapDerivationReference {
	fingerprint = strings.TrimSpace(fingerprint)
	encoded := strings.TrimPrefix(fingerprint, fingerprintVersionV2+"-")
	basePath := filepath.Join(stems.BasePath, "roots", fingerprintVersionV2, encoded)
	var heapDerivationRef = UnsafeHeapDerivationReference{
		BasePath:          basePath,
		SourceFingerprint: fingerprint,
		Derivations:       stems,
	}
	return &heapDerivationRef
}

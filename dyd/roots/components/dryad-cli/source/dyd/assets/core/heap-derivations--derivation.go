package core

import (
	"strings"
)

func (stems *SafeHeapDerivationsReference) Derivation(fingerprint string) *UnsafeHeapDerivationReference {
	fingerprint = strings.TrimSpace(fingerprint)
	var heapDerivationRef = UnsafeHeapDerivationReference{
		BasePath:          "",
		SourceFingerprint: fingerprint,
		Derivations:       stems,
	}
	return &heapDerivationRef
}

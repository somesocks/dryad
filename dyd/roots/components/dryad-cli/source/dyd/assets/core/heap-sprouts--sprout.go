package core

import (
	"path/filepath"
	"strings"
)

func (sprouts *SafeHeapSproutsReference) Sprout(fingerprint string) *UnsafeHeapSproutReference {
	fingerprint = strings.TrimSpace(fingerprint)
	encoded := strings.TrimPrefix(fingerprint, fingerprintVersionV2+"-")
	basePath := filepath.Join(sprouts.BasePath, fingerprintVersionV2, encoded)
	var heapSproutRef = UnsafeHeapSproutReference{
		BasePath:    basePath,
		Fingerprint: fingerprint,
		Sprouts:     sprouts,
	}
	return &heapSproutRef
}

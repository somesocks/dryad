package core

import (
	"dryad/internal/filepath"
	// "dryad/task"
	// zlog "github.com/rs/zerolog/log"
)

func (traits *SafeSproutTraitsReference) Trait(path string) *UnsafeSproutTraitReference {
	var sproutTraitRef = UnsafeSproutTraitReference{
		BasePath: filepath.Join(traits.BasePath, path),
		Traits:   traits,
	}
	return &sproutTraitRef
}

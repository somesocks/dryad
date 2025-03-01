package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

func (sprout *SafeSproutReference) Traits() (*UnsafeSproutTraitsReference) {
	var sproutTraitsRef = UnsafeSproutTraitsReference{
		BasePath: filepath.Join(sprout.BasePath, "dyd", "traits"),
		Sprout: sprout,
	}
	return &sproutTraitsRef
}
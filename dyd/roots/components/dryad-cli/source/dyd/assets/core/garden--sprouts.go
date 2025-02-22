package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (sg *SafeGardenReference) Sprouts() (*UnsafeSproutsReference) {
	var ref = UnsafeSproutsReference{
		BasePath: filepath.Join(sg.BasePath, "dyd", "sprouts"),
		Garden: sg,
	}
	return &ref
}
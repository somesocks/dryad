package core

import (
	"dryad/task"
	"path/filepath"
	// zlog "github.com/rs/zerolog/log"
)

type SproutRunRequest struct {
	WorkingPath  string
	MainOverride string
	Context      string
	Env          map[string]string
	Args         []string
	JoinStdout   bool
	LogStdout    struct {
		Path string
		Name string
	}
	JoinStderr bool
	LogStderr  struct {
		Path string
		Name string
	}
	InheritEnv bool
}

func (sprout *SafeSproutReference) Run(
	ctx *task.ExecutionContext,
	req SproutRunRequest,
) error {
	stemPath, err := StemPath(sprout.BasePath)
	if err != nil {
		return err
	}

	logStdout := req.LogStdout
	logStderr := req.LogStderr

	if logStdout.Path != "" && logStdout.Name == "" {
		relSproutPath, relErr := filepath.Rel(
			sprout.Sprouts.Garden.BasePath,
			sprout.BasePath,
		)
		if relErr != nil {
			return relErr
		}
		logStdout.Name = "dyd-sprout-run--" +
			sanitizePathSegment(relSproutPath) +
			".out"
	}

	if logStderr.Path != "" && logStderr.Name == "" {
		relSproutPath, relErr := filepath.Rel(
			sprout.Sprouts.Garden.BasePath,
			sprout.BasePath,
		)
		if relErr != nil {
			return relErr
		}
		logStderr.Name = "dyd-sprout-run--" +
			sanitizePathSegment(relSproutPath) +
			".err"
	}

	err = StemRun(
		StemRunRequest{
			Garden:       sprout.Sprouts.Garden,
			StemPath:     stemPath,
			WorkingPath:  req.WorkingPath,
			MainOverride: req.MainOverride,
			Context:      req.Context,
			Env:          req.Env,
			Args:         req.Args,
			JoinStdout:   req.JoinStdout,
			LogStdout:    logStdout,
			JoinStderr:   req.JoinStderr,
			LogStderr:    logStderr,
			InheritEnv:   req.InheritEnv,
		},
	)

	return err
}

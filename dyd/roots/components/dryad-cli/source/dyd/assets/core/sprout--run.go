
package core

import (
	"dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

type SproutRunRequest struct {
	WorkingPath  string
	MainOverride string
	Context      string
	Env          map[string]string
	Args         []string
	JoinStdout   bool
	LogStdout    string
	JoinStderr   bool
	LogStderr    string
	InheritEnv   bool
}

func (sprout *SafeSproutReference) Run(
	ctx * task.ExecutionContext,
	req SproutRunRequest,
) (error) {

	err := StemRun(
		StemRunRequest{
			Garden: sprout.Sprouts.Garden,
			StemPath: sprout.BasePath,
			WorkingPath: req.WorkingPath,
			MainOverride: req.MainOverride,
			Context: req.Context,
			Env: req.Env,
			Args: req.Args,
			JoinStdout: req.JoinStdout,
			LogStdout: req.LogStdout,
			JoinStderr: req.JoinStderr,
			LogStderr: req.LogStderr,
			InheritEnv: req.InheritEnv,						
		},
	)

	return err
}

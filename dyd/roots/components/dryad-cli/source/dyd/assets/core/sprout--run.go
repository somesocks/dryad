
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
	JoinStderr   bool
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
			JoinStderr: req.JoinStderr,
			InheritEnv: req.InheritEnv,						
		},
	)

	return err
}

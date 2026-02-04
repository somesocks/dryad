package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	zlog "github.com/rs/zerolog/log"
)

// stage 0 - build a shallow partial clone of the root into a working directory,
// so we can build it into a stem
func rootDevelop_stage0(ctx *task.ExecutionContext, rootPath string, workspacePath string) error {
	// fmt.Println("rootDevelop_stage0 ", rootPath, " ", workspacePath)

	rootPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(
		filepath.Join(workspacePath, "dyd"),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	exists, err := fileExists(filepath.Join(rootPath, "dyd", "assets"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "assets"),
			filepath.Join(workspacePath, "dyd", "assets"),
			rootDevelopCopyOptions{ApplyIgnore: true},
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "commands"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "commands"),
			filepath.Join(workspacePath, "dyd", "commands"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "docs"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "docs"),
			filepath.Join(workspacePath, "dyd", "docs"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "traits"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "traits"),
			filepath.Join(workspacePath, "dyd", "traits"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "secrets"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "secrets"),
			filepath.Join(workspacePath, "dyd", "secrets"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return err
		}
	}

	err = os.Mkdir(filepath.Join(workspacePath, "dyd", "dependencies"), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// stage 1 - walk through the root dependencies,
// and add the fingerprint as a dependency
func rootDevelop_stage1(
	ctx *task.ExecutionContext,
	rootPath string,
	workspacePath string,
	roots *SafeRootsReference,
) error {
	
	rootRef := SafeRootReference{
		BasePath: rootPath,
		Roots: roots,
	}

	err, requirementsRef := rootRef.Requirements().Resolve(ctx)
	if err != nil {
		return err
	}

	err = requirementsRef.Walk(task.SERIAL_CONTEXT, RootRequirementsWalkRequest{
		OnMatch: func (ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
			err, safeDepReference := requirement.Target(ctx)
			if err != nil {
				return err, nil
			}

			err, dependencyFingerprint := safeDepReference.Build(
				ctx,
				RootBuildRequest{},
			)
			if err != nil {
				return err, nil
			}
	
			dependencyHeapPath := filepath.Join(requirement.Requirements.Root.Roots.Garden.BasePath, "dyd", "heap", "stems", dependencyFingerprint)
	
			dependencyName := filepath.Base(requirement.BasePath)
		
			targetDepPath := filepath.Join(workspacePath, "dyd", "dependencies", dependencyName)
		
			err = os.Symlink(dependencyHeapPath, targetDepPath)

			return err, nil
		},
	});
	if err != nil {
		return err
	}

	return nil
}

// stage 2 - generate the artificial links to all executable stems for the path
func rootDevelop_stage2(workspacePath string) error {
	err := rootBuild_pathPrepare(workspacePath)
	if err != nil {
		return err
	}
	err = rootBuild_requirementsPrepare(workspacePath)
	if err != nil {
		return err
	}
	return nil
}

// stage 3 - finalize the stem by generating fingerprints,
func rootDevelop_stage3(rootPath string, workspacePath string) (string, error) {
	// fmt.Println("rootDevelop_stage3 ", rootPath)

	err, stemFingerprint := stemFinalize(task.SERIAL_CONTEXT, workspacePath)
	return stemFingerprint, err
}

// stage 4 - execute the `dyd-root-develop-start` command if it exists
func rootDevelop_stage4(
	rootStemPath string,
	stemBuildPath string,
	rootFingerprint string,
	garden *SafeGardenReference,
	editor string,
	editorArgs []string,
	inherit bool,
	devSocket string,
) error {
	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return err
	}

	onDevelopStartPath := filepath.Join(rootStemPath, "dyd", "commands", "dyd-root-develop-start")

	onDevelopStartExists, err := fileExists(onDevelopStartPath)
	if err != nil {
		return err
	}

	if !onDevelopStartExists {
		return nil
	}

	env := map[string]string{
		"DYD_BUILD": stemBuildPath,
	}
	if devSocket != "" {
		env["DYD_DEV_SOCKET"] = devSocket
	}

	err = StemRun(StemRunRequest{
		Garden: garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: onDevelopStartPath,
		Env:         env,
		Args:       editorArgs,
		JoinStdout: true,
		JoinStderr: true,
		InheritEnv: inherit,
	})

	return err

}

// stage 6 - execute the `dyd-root-develop-stop` command if it exists
func rootDevelop_stage6(
	rootStemPath string,
	stemBuildPath string,
	rootFingerprint string,
	garden *SafeGardenReference,
	editor string,
	editorArgs []string,
	inherit bool,
	devSocket string,
) error {
	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return err
	}

	onDevelopStopPath := filepath.Join(rootStemPath, "dyd", "commands", "dyd-root-develop-stop")

	onDevelopStopExists, err := fileExists(onDevelopStopPath)
	if err != nil {
		return err
	}

	if !onDevelopStopExists {
		return nil
	}

	env := map[string]string{
		"DYD_BUILD": stemBuildPath,
	}
	if devSocket != "" {
		env["DYD_DEV_SOCKET"] = devSocket
	}

	err = StemRun(StemRunRequest{
		Garden:   garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: onDevelopStopPath,
		Env:         env,
		Args:       editorArgs,
		JoinStdout: true,
		JoinStderr: true,
		InheritEnv: inherit,
	})

	return err

}

type rootDevelopEditorProcess struct {
	mu            sync.Mutex
	cmd           *exec.Cmd
	stopRequested bool
}

func (proc *rootDevelopEditorProcess) setCmd(cmd *exec.Cmd) {
	if proc == nil {
		return
	}
	proc.mu.Lock()
	defer proc.mu.Unlock()
	proc.cmd = cmd
}

func (proc *rootDevelopEditorProcess) clearCmd() {
	if proc == nil {
		return
	}
	proc.mu.Lock()
	defer proc.mu.Unlock()
	proc.cmd = nil
}

func (proc *rootDevelopEditorProcess) wasStopRequested() bool {
	if proc == nil {
		return false
	}
	proc.mu.Lock()
	defer proc.mu.Unlock()
	return proc.stopRequested
}

func (proc *rootDevelopEditorProcess) requestStop() error {
	if proc == nil {
		return nil
	}
	proc.mu.Lock()
	proc.stopRequested = true
	cmd := proc.cmd
	proc.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	err := cmd.Process.Signal(os.Interrupt)
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	return nil
}

// stage 5 - execute the editor in the root to build its stem
func rootDevelop_stage5(
	rootStemPath string,
	stemBuildPath string,
	rootFingerprint string,
	garden *SafeGardenReference,
	editor string,
	editorArgs []string,
	inherit bool,
	devSocket string,
	editorProcess *rootDevelopEditorProcess,
) (string, error) {
	// fmt.Println("rootDevelop_stage5 ", rootStemPath, stemBuildPath)

	// find default development editor if not passed in
	// fallback to 'sh' if no dyd-root-develop command exists
	if editor == "" {

		onDevelopPath := filepath.Join(rootStemPath, "dyd", "commands", "dyd-root-develop")
		onDevelopExists, err := fileExists(onDevelopPath)
		if err != nil {
			return "", err
		}

		if onDevelopExists {
			editor = onDevelopPath
		} else {
			editor = "sh"
		}

	}

	if editorProcess != nil && editorProcess.wasStopRequested() {
		return "", nil
	}

	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return "", err
	}

	env := map[string]string{
		"DYD_BUILD": stemBuildPath,
	}
	if devSocket != "" {
		env["DYD_DEV_SOCKET"] = devSocket
	}

	cmd, err := StemRunCommand(StemRunRequest{
		Garden: garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: editor,
		Env:         env,
		Args:       editorArgs,
		JoinStdout: true,
		JoinStderr: true,
		InheritEnv: inherit,
	})
	if err != nil {
		return "", err
	}

	if editorProcess != nil {
		editorProcess.setCmd(cmd)
	}
	if err := cmd.Start(); err != nil {
		if editorProcess != nil {
			editorProcess.clearCmd()
		}
		return "", err
	}

	err = cmd.Wait()
	if editorProcess != nil {
		editorProcess.clearCmd()
		if err != nil && editorProcess.wasStopRequested() {
			return "", nil
		}
	}

	return "", err
}

type rootDevelopRequest struct {
	Root *SafeRootReference
	Editor string
	EditorArgs []string
	Inherit bool
}

func rootDevelop(
	ctx *task.ExecutionContext,
	req rootDevelopRequest,
) (string, error) {

	gardenPath := req.Root.Roots.Garden.BasePath
	editor := req.Editor
	editorArgs := req.EditorArgs
	inherit := req.Inherit

	rootPath := req.Root.BasePath

	absRootPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", err
	}

	relRootPath, err := filepath.Rel(
		filepath.Join(gardenPath, "dyd", "roots"),
		absRootPath,
	)
	if err != nil {
		return "", err
	}

	zlog.Info().Msg("creating development environment for root " + relRootPath)

	// prepare a workspace
	workspacePath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return "", err
	}
	defer dydfs.RemoveAll(task.SERIAL_CONTEXT, workspacePath)

	err = rootDevelop_stage0(ctx, rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	snapshot, err := rootDevelop_collectAll(ctx, rootPath)
	if err != nil {
		return "", err
	}

	err = rootDevelop_stage1(ctx, rootPath, workspacePath, req.Root.Roots)
	if err != nil {
		return "", err
	}

	err = rootDevelop_stage2(workspacePath)
	if err != nil {
		return "", err
	}

	rootFingerprint, err := rootDevelop_stage3(rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	editorProcess := &rootDevelopEditorProcess{}

	devSocketPath := filepath.Join(workspacePath, "dev.sock")
	devServer, err := rootDevelopIPC_start(devSocketPath, rootDevelopIPCHandlers{
		OnSave: func() error {
			conflicts, err := rootDevelop_saveChanges(ctx, rootPath, workspacePath, snapshot)
			if err != nil {
				return err
			}
			if len(conflicts) > 0 {
				return fmt.Errorf("root develop save: %d conflicts", len(conflicts))
			}
			return nil
		},
		OnStatus: func() ([]string, []string, error) {
			return rootDevelop_statusChanges(ctx, rootPath, workspacePath, snapshot)
		},
		OnStop: func() error {
			return editorProcess.requestStop()
		},
	})
	if err != nil {
		return "", err
	}
	defer devServer.Close()

	// otherwise run the root in a build env
	stemBuildPath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return "", err
	}
	defer dydfs.RemoveAll(ctx, stemBuildPath)

	err = rootDevelop_stage4(
		workspacePath,
		stemBuildPath,
		rootFingerprint,
		req.Root.Roots.Garden,
		editor,
		editorArgs,
		inherit,
		devSocketPath,
	)
	if err != nil {
		return "", err
	}

	stemBuildFingerprint, onDevelopErr := rootDevelop_stage5(
		workspacePath,
		stemBuildPath,
		rootFingerprint,
		req.Root.Roots.Garden,
		editor,
		editorArgs,
		inherit,
		devSocketPath,
		editorProcess,
	)

	onStopErr := rootDevelop_stage6(
		workspacePath,
		stemBuildPath,
		rootFingerprint,
		req.Root.Roots.Garden,
		editor,
		editorArgs,
		inherit,
		devSocketPath,
	)

	if onDevelopErr != nil {
		return "", onDevelopErr
	} else if onStopErr != nil {
		return "", onStopErr
	} else {
		return stemBuildFingerprint, nil
	}

}

type RootDevelopRequest struct {
	Editor string
	EditorArgs []string
	Inherit bool
}

func (root *SafeRootReference) Develop(ctx *task.ExecutionContext, req RootDevelopRequest) (error, string) {
	res, err := rootDevelop(
		ctx,
		rootDevelopRequest{
			Root: root,
			Editor: req.Editor,
			EditorArgs: req.EditorArgs,
			Inherit: req.Inherit,
		},
	)
	return err, res
}

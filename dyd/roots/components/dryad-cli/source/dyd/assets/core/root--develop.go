package core

import (
	"bufio"
	dydfs "dryad/filesystem"
	"dryad/task"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/mattn/go-isatty"
	zlog "github.com/rs/zerolog/log"
)

// stage 0 - build a shallow partial clone of the root into a working directory,
// so we can build it into a stem
func rootDevelop_stage0(ctx *task.ExecutionContext, snapshotStemPath string, workspacePath string) error {
	// fmt.Println("rootDevelop_stage0 ", rootPath, " ", workspacePath)

	snapshotStemPath, err := filepath.EvalSymlinks(snapshotStemPath)
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

	exists, err := fileExists(filepath.Join(snapshotStemPath, "dyd", "assets"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "assets"),
			filepath.Join(workspacePath, "dyd", "assets"),
			"dyd/assets",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "commands"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "commands"),
			filepath.Join(workspacePath, "dyd", "commands"),
			"dyd/commands",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "docs"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "docs"),
			filepath.Join(workspacePath, "dyd", "docs"),
			"dyd/docs",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "traits"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "traits"),
			filepath.Join(workspacePath, "dyd", "traits"),
			"dyd/traits",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "secrets"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "secrets"),
			filepath.Join(workspacePath, "dyd", "secrets"),
			"dyd/secrets",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "requirements"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "requirements"),
			filepath.Join(workspacePath, "dyd", "requirements"),
			"dyd/requirements",
		)
		if err != nil {
			return err
		}
	} else {
		err = os.MkdirAll(filepath.Join(workspacePath, "dyd", "requirements"), os.ModePerm)
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

func rootDevelop_resetWorkspace(
	ctx *task.ExecutionContext,
	snapshotStemPath string,
	workspacePath string,
) error {
	snapshotStemPath, err := filepath.EvalSymlinks(snapshotStemPath)
	if err != nil {
		return err
	}

	exists, err := fileExists(filepath.Join(snapshotStemPath, "dyd", "assets"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "assets"),
			filepath.Join(workspacePath, "dyd", "assets"),
			"dyd/assets",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "commands"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "commands"),
			filepath.Join(workspacePath, "dyd", "commands"),
			"dyd/commands",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "docs"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "docs"),
			filepath.Join(workspacePath, "dyd", "docs"),
			"dyd/docs",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "traits"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "traits"),
			filepath.Join(workspacePath, "dyd", "traits"),
			"dyd/traits",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "secrets"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "secrets"),
			filepath.Join(workspacePath, "dyd", "secrets"),
			"dyd/secrets",
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(snapshotStemPath, "dyd", "requirements"))
	if err != nil {
		return err
	}
	if exists {
		err = rootDevelop_copyDirFromStem(
			ctx,
			filepath.Join(snapshotStemPath, "dyd", "requirements"),
			filepath.Join(workspacePath, "dyd", "requirements"),
			"dyd/requirements",
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func rootDevelop_devDir(workspacePath string) string {
	return filepath.Join(workspacePath, ".dyd-develop")
}

func rootDevelop_snapshotFile(workspacePath string) string {
	return filepath.Join(rootDevelop_devDir(workspacePath), "snapshot-stem")
}

func rootDevelop_snapshotStemPath(
	garden *SafeGardenReference,
	workspacePath string,
) (string, error) {
	bytes, err := os.ReadFile(rootDevelop_snapshotFile(workspacePath))
	if err != nil {
		return "", err
	}
	fingerprint := strings.TrimSpace(string(bytes))
	return filepath.Join(garden.BasePath, "dyd", "heap", "stems", fingerprint), nil
}

func rootDevelop_createSnapshotStem(
	ctx *task.ExecutionContext,
	rootPath string,
	garden *SafeGardenReference,
) (string, error) {
	snapshotWorkspace, err := os.MkdirTemp("", "dryad-snapshot-*")
	if err != nil {
		return "", err
	}
	defer dydfs.RemoveAll(task.SERIAL_CONTEXT, snapshotWorkspace)

	err = os.MkdirAll(filepath.Join(snapshotWorkspace, "dyd"), os.ModePerm)
	if err != nil {
		return "", err
	}

	exists, err := fileExists(filepath.Join(rootPath, "dyd", "assets"))
	if err != nil {
		return "", err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "assets"),
			filepath.Join(snapshotWorkspace, "dyd", "assets"),
			rootDevelopCopyOptions{ApplyIgnore: true},
		)
		if err != nil {
			return "", err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "commands"))
	if err != nil {
		return "", err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "commands"),
			filepath.Join(snapshotWorkspace, "dyd", "commands"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return "", err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "docs"))
	if err != nil {
		return "", err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "docs"),
			filepath.Join(snapshotWorkspace, "dyd", "docs"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return "", err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "traits"))
	if err != nil {
		return "", err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "traits"),
			filepath.Join(snapshotWorkspace, "dyd", "traits"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return "", err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "secrets"))
	if err != nil {
		return "", err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "secrets"),
			filepath.Join(snapshotWorkspace, "dyd", "secrets"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return "", err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "requirements"))
	if err != nil {
		return "", err
	}
	if exists {
		err = rootDevelop_copyDir(
			ctx,
			filepath.Join(rootPath, "dyd", "requirements"),
			filepath.Join(snapshotWorkspace, "dyd", "requirements"),
			rootDevelopCopyOptions{},
		)
		if err != nil {
			return "", err
		}
	} else {
		err = os.MkdirAll(filepath.Join(snapshotWorkspace, "dyd", "requirements"), os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	err = os.MkdirAll(filepath.Join(snapshotWorkspace, "dyd", "dependencies"), os.ModePerm)
	if err != nil {
		return "", err
	}

	err, snapshotFingerprint := stemFinalize(ctx, snapshotWorkspace)
	if err != nil {
		return "", err
	}

	err, heap := garden.Heap().Resolve(ctx)
	if err != nil {
		return "", err
	}

	err, stems := heap.Stems().Resolve(ctx)
	if err != nil {
		return "", err
	}

	err, _ = stems.AddStem(
		ctx,
		HeapAddStemRequest{
			StemPath: snapshotWorkspace,
		},
	)
	if err != nil {
		return "", err
	}

	return snapshotFingerprint, nil
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
		Roots:    roots,
	}

	err, requirementsRef := rootRef.Requirements().Resolve(ctx)
	if err != nil {
		return err
	}

	err = requirementsRef.Walk(task.SERIAL_CONTEXT, RootRequirementsWalkRequest{
		OnMatch: func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
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
	})
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
	return nil
}

// stage 3 - finalize the stem by generating fingerprints,
func rootDevelop_stage3(rootPath string, workspacePath string) (string, error) {
	// fmt.Println("rootDevelop_stage3 ", rootPath)

	err, stemFingerprint := stemFinalize(task.SERIAL_CONTEXT, workspacePath)
	return stemFingerprint, err
}

type rootDevelopShellProcess struct {
	mu            sync.Mutex
	cmd           *exec.Cmd
	stopRequested bool
}

func (proc *rootDevelopShellProcess) setCmd(cmd *exec.Cmd) {
	if proc == nil {
		return
	}
	proc.mu.Lock()
	defer proc.mu.Unlock()
	proc.cmd = cmd
}

func (proc *rootDevelopShellProcess) clearCmd() {
	if proc == nil {
		return
	}
	proc.mu.Lock()
	defer proc.mu.Unlock()
	proc.cmd = nil
}

func (proc *rootDevelopShellProcess) wasStopRequested() bool {
	if proc == nil {
		return false
	}
	proc.mu.Lock()
	defer proc.mu.Unlock()
	return proc.stopRequested
}

func (proc *rootDevelopShellProcess) requestStop() error {
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

// stage 5 - execute the shell command in the root development environment
func rootDevelop_stage5(
	rootStemPath string,
	stemBuildPath string,
	rootFingerprint string,
	garden *SafeGardenReference,
	shell string,
	shellArgs []string,
	inherit bool,
	devSocket string,
	shellProcess *rootDevelopShellProcess,
) (string, error) {
	// fmt.Println("rootDevelop_stage5 ", rootStemPath, stemBuildPath)
	if shell == "" {
		shell = "sh"
	}
	if len(shellArgs) > 0 {
		// Preserve passthrough arg boundaries when routing through shell execution.
		shellArgs = append([]string{"-c", "exec \"$@\"", shell}, shellArgs...)
	}

	if shellProcess != nil && shellProcess.wasStopRequested() {
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

	instance, err := StemRunCommand(StemRunRequest{
		Garden:       garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: shell,
		Env:          env,
		Args:         shellArgs,
		JoinStdout:   true,
		JoinStderr:   true,
		InheritEnv:   inherit,
	})
	if err != nil {
		return "", err
	}
	if instance.Close != nil {
		defer instance.Close()
	}

	if shellProcess != nil {
		shellProcess.setCmd(instance.Cmd)
	}
	if err := instance.Cmd.Start(); err != nil {
		if shellProcess != nil {
			shellProcess.clearCmd()
		}
		return "", err
	}

	err = instance.Cmd.Wait()
	if shellProcess != nil {
		shellProcess.clearCmd()
		if err != nil && shellProcess.wasStopRequested() {
			return "", nil
		}
	}

	return "", err
}

func rootDevelop_handleUnsavedChanges(
	ctx *task.ExecutionContext,
	rootPath string,
	workspacePath string,
	garden *SafeGardenReference,
	onExit string,
) error {
	currentSnapshotPath, err := rootDevelop_snapshotStemPath(garden, workspacePath)
	if err != nil {
		return err
	}

	entries, err := rootDevelop_statusChanges(ctx, rootPath, workspacePath, currentSnapshotPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	snapshotFingerprint, err := rootDevelop_createSnapshotStem(ctx, workspacePath, garden)
	if err != nil {
		return err
	}

	switch strings.TrimSpace(strings.ToLower(onExit)) {
	case "":
	case "ask":
	case "save":
		conflicts, err := rootDevelop_saveChanges(ctx, rootPath, workspacePath, currentSnapshotPath)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return fmt.Errorf("root develop save: %d conflicts", len(conflicts))
		}
		return nil
	case "discard":
		fmt.Fprintf(os.Stderr, "warning: discarded unsaved changes; snapshot %s\n", snapshotFingerprint)
		return nil
	default:
		return fmt.Errorf("invalid on-exit action: %s", onExit)
	}

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "warning: root develop exited with unsaved changes; snapshot %s\n", snapshotFingerprint)
		return nil
	}

	fmt.Fprintln(os.Stderr, "unsaved changes:")
	for _, entry := range entries {
		fmt.Fprintln(os.Stderr, entry.Code+" "+entry.Path)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "save changes? [s=save, d=discard]: ")
		line, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, syscall.EIO) {
			return readErr
		}

		choice := strings.TrimSpace(strings.ToLower(line))
		if readErr != nil && (errors.Is(readErr, io.EOF) || errors.Is(readErr, syscall.EIO)) && choice == "" {
			fmt.Fprintf(os.Stderr, "warning: root develop exited with unsaved changes; snapshot %s\n", snapshotFingerprint)
			return nil
		}

		switch choice {
		case "s", "save", "y", "yes":
			conflicts, err := rootDevelop_saveChanges(ctx, rootPath, workspacePath, currentSnapshotPath)
			if err != nil {
				return err
			}
			if len(conflicts) > 0 {
				fmt.Fprintf(os.Stderr, "warning: save reported %d conflicts\n", len(conflicts))
				continue
			}
			return nil
		case "d", "discard", "n", "no":
			fmt.Fprintf(os.Stderr, "warning: discarded unsaved changes; snapshot %s\n", snapshotFingerprint)
			return nil
		default:
			fmt.Fprintln(os.Stderr, "enter 's' to save or 'd' to discard")
		}
	}
}

type rootDevelopRequest struct {
	Root      *SafeRootReference
	Shell     string
	ShellArgs []string
	Inherit   bool
	OnExit    string
}

func rootDevelop(
	ctx *task.ExecutionContext,
	req rootDevelopRequest,
) (string, error) {

	gardenPath := req.Root.Roots.Garden.BasePath
	shell := req.Shell
	shellArgs := req.ShellArgs
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

	devDir := rootDevelop_devDir(workspacePath)
	err = os.MkdirAll(devDir, 0o755)
	if err != nil {
		return "", err
	}

	snapshotFingerprint, err := rootDevelop_createSnapshotStem(ctx, rootPath, req.Root.Roots.Garden)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(rootDevelop_snapshotFile(workspacePath), []byte(snapshotFingerprint), 0o644)
	if err != nil {
		return "", err
	}

	snapshotStemPath := filepath.Join(req.Root.Roots.Garden.BasePath, "dyd", "heap", "stems", snapshotFingerprint)

	err = rootDevelop_stage0(ctx, snapshotStemPath, workspacePath)
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

	shellProcess := &rootDevelopShellProcess{}

	devSocketPath := filepath.Join(devDir, "host.sock")
	devServer, err := rootDevelopIPC_start(devSocketPath, rootDevelopIPCHandlers{
		OnSave: func() error {
			currentSnapshotPath, err := rootDevelop_snapshotStemPath(req.Root.Roots.Garden, workspacePath)
			if err != nil {
				return err
			}
			conflicts, err := rootDevelop_saveChanges(ctx, rootPath, workspacePath, currentSnapshotPath)
			if err != nil {
				return err
			}
			if len(conflicts) > 0 {
				return fmt.Errorf("root develop save: %d conflicts", len(conflicts))
			}
			return nil
		},
		OnStatus: func() ([]rootDevelopStatusEntry, error) {
			currentSnapshotPath, err := rootDevelop_snapshotStemPath(req.Root.Roots.Garden, workspacePath)
			if err != nil {
				return nil, err
			}
			return rootDevelop_statusChanges(ctx, rootPath, workspacePath, currentSnapshotPath)
		},
		OnSnapshot: func() (string, error) {
			fingerprint, err := rootDevelop_createSnapshotStem(ctx, workspacePath, req.Root.Roots.Garden)
			if err != nil {
				return "", err
			}
			if err := os.WriteFile(rootDevelop_snapshotFile(workspacePath), []byte(fingerprint), 0o644); err != nil {
				return "", err
			}
			return fingerprint, nil
		},
		OnReset: func() error {
			currentSnapshotPath, err := rootDevelop_snapshotStemPath(req.Root.Roots.Garden, workspacePath)
			if err != nil {
				return err
			}
			return rootDevelop_resetWorkspace(ctx, currentSnapshotPath, workspacePath)
		},
		OnStop: func() error {
			return shellProcess.requestStop()
		},
	})
	if err != nil {
		return "", err
	}
	defer devServer.Close()

	// use a disposable build directory at the root of the development workspace.
	stemBuildPath := filepath.Join(workspacePath, "out")

	stemBuildFingerprint, onDevelopErr := rootDevelop_stage5(
		workspacePath,
		stemBuildPath,
		rootFingerprint,
		req.Root.Roots.Garden,
		shell,
		shellArgs,
		inherit,
		devSocketPath,
		shellProcess,
	)

	promptErr := rootDevelop_handleUnsavedChanges(ctx, rootPath, workspacePath, req.Root.Roots.Garden, req.OnExit)
	if promptErr != nil {
		return "", promptErr
	}

	if onDevelopErr != nil {
		return "", onDevelopErr
	} else {
		return stemBuildFingerprint, nil
	}

}

type RootDevelopRequest struct {
	Shell     string
	ShellArgs []string
	Inherit   bool
	OnExit    string
}

func (root *SafeRootReference) Develop(ctx *task.ExecutionContext, req RootDevelopRequest) (error, string) {
	res, err := rootDevelop(
		ctx,
		rootDevelopRequest{
			Root:      root,
			Shell:     req.Shell,
			ShellArgs: req.ShellArgs,
			Inherit:   req.Inherit,
			OnExit:    req.OnExit,
		},
	)
	return err, res
}

package core

import (

	dydfs "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

// stage 0 - build a shallow partial clone of the root into a working directory,
// so we can build it into a stem
func rootDevelop_stage0(rootPath string, workspacePath string) error {
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
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "assets"),
			filepath.Join(workspacePath, "dyd", "assets"),
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
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "commands"),
			filepath.Join(workspacePath, "dyd", "commands"),
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
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "docs"),
			filepath.Join(workspacePath, "dyd", "docs"),
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
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "traits"),
			filepath.Join(workspacePath, "dyd", "traits"),
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
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "secrets"),
			filepath.Join(workspacePath, "dyd", "secrets"),
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
	// fmt.Println("rootDevelop_stage1 ", rootPath, " ", workspacePath)

	// walk through the dependencies, build them, and add the fingerprint as a dependency
	rootsPath := filepath.Join(rootPath, "dyd", "requirements")

	dependencies, err := filepath.Glob(filepath.Join(rootsPath, "*"))
	if err != nil {
		return err
	}

	for _, dependencyPath := range dependencies {
		var unsafeDepReference = UnsafeRootReference{
			Roots: roots,
			BasePath: dependencyPath,
		}
		var safeDepReference SafeRootReference
		var err error

		err, safeDepReference = unsafeDepReference.Resolve(ctx, nil)
		if err != nil {
			return err
		}

		err, _ = safeDepReference.Build(
			ctx,
			RootBuildRequest{
			},
		)
		if err != nil {
			return err
		}

		dependencyName := filepath.Base(dependencyPath)

		// fmt.Println("[trace] RootBuild gardenPath", gardenPath)

		absRootPath, err := filepath.EvalSymlinks(dependencyPath)
		if err != nil {
			return err
		}

		relRootPath, err := filepath.Rel(
			roots.BasePath,
			absRootPath,
		)
		if err != nil {
			return err
		}

		sproutPath := filepath.Join(roots.Garden.BasePath, "dyd", "sprouts", relRootPath)
		targetDepPath := filepath.Join(workspacePath, "dyd", "dependencies", dependencyName)

		err = os.Symlink(sproutPath, targetDepPath)

		if err != nil {
			return err
		}
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

// stage 4 - execute the `on-develop-start` command if it exists
func rootDevelop_stage4(
	rootStemPath string,
	stemBuildPath string,
	rootFingerprint string,
	garden *SafeGardenReference,
	editor string,
	editorArgs []string,
	inherit bool,
) error {
	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return err
	}

	onDevelopStartPath := filepath.Join(rootStemPath, "dyd", "commands", "on-develop-start")

	onDevelopStartExists, err := fileExists(onDevelopStartPath)
	if err != nil {
		return err
	}

	if !onDevelopStartExists {
		return nil
	}

	err = StemRun(StemRunRequest{
		Garden: garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: onDevelopStartPath,
		Env: map[string]string{
			"DYD_BUILD": stemBuildPath,
		},
		Args:       editorArgs,
		JoinStdout: true,
		JoinStderr: true,
		InheritEnv: inherit,
	})

	return err

}

// stage 6 - execute the `on-develop-stop` command if it exists
func rootDevelop_stage6(
	rootStemPath string,
	stemBuildPath string,
	rootFingerprint string,
	garden *SafeGardenReference,
	editor string,
	editorArgs []string,
	inherit bool,
) error {
	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return err
	}

	onDevelopStopPath := filepath.Join(rootStemPath, "dyd", "commands", "on-develop-stop")

	onDevelopStopExists, err := fileExists(onDevelopStopPath)
	if err != nil {
		return err
	}

	if !onDevelopStopExists {
		return nil
	}

	err = StemRun(StemRunRequest{
		Garden:   garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: onDevelopStopPath,
		Env: map[string]string{
			"DYD_BUILD": stemBuildPath,
		},
		Args:       editorArgs,
		JoinStdout: true,
		JoinStderr: true,
		InheritEnv: inherit,
	})

	return err

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
) (string, error) {
	// fmt.Println("rootDevelop_stage5 ", rootStemPath, stemBuildPath)

	// find default development editor if not passed in
	// fallback to 'sh' if no on-develop command exists
	if editor == "" {

		onDevelopPath := filepath.Join(rootStemPath, "dyd", "commands", "on-develop")
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

	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return "", err
	}

	err = StemRun(StemRunRequest{
		Garden: garden,
		StemPath:     rootStemPath,
		WorkingPath:  rootStemPath,
		MainOverride: editor,
		Env: map[string]string{
			"DYD_BUILD": stemBuildPath,
		},
		Args:       editorArgs,
		JoinStdout: true,
		JoinStderr: true,
		InheritEnv: inherit,
	})

	return "", err
}

type RootDevelopRequest struct {
	Root *SafeRootReference
	Editor string
	EditorArgs []string
	Inherit bool
}

func RootDevelop(
	ctx *task.ExecutionContext,
	req RootDevelopRequest,
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

	err = rootDevelop_stage0(rootPath, workspacePath)
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
	)

	onStopErr := rootDevelop_stage6(
		workspacePath,
		stemBuildPath,
		rootFingerprint,
		req.Root.Roots.Garden,
		editor,
		editorArgs,
		inherit,
	)

	if onDevelopErr != nil {
		return "", onDevelopErr
	} else if onStopErr != nil {
		return "", onStopErr
	} else {
		return stemBuildFingerprint, nil
	}

}

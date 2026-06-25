package core

import (
	"dryad/internal/filepath"
	"dryad/task"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootRequirementCanonicalEnvName(t *testing.T) {
	assert := assert.New(t)

	err, name := rootRequirementCanonicalEnvName("my-display.name")
	assert.Nil(err)
	assert.Equal("MY_DISPLAY_NAME", name)

	err, _ = rootRequirementCanonicalEnvName("1display")
	assert.NotNil(err)
}

func TestRootBuildStage1_MaterializesEnvRequirementWithFingerprint(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	rootPath := filepath.Join(gardenPath, "dyd", "roots", "root-01")
	workspacePath := t.TempDir()
	requirementsPath := filepath.Join(rootPath, "dyd", "requirements")

	writeFileForTest(t, filepath.Join(requirementsPath, "my-display"), "env:display")
	writeFileForTest(t, filepath.Join(workspacePath, "dyd", ".keep"), "")
	t.Setenv("DISPLAY", ":99")

	err, results := rootBuild_stage1(
		task.SERIAL_CONTEXT,
		rootBuild_stage1_request{
			Roots: &SafeRootsReference{
				BasePath: filepath.Join(gardenPath, "dyd", "roots"),
				Garden:   &SafeGardenReference{BasePath: gardenPath},
			},
			RootPath:                 rootPath,
			WorkspacePath:            workspacePath,
			SelectedRequirementsPath: requirementsPath,
		},
	)
	assert.Nil(err)
	assert.Empty(results)

	materialized := readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "requirements", "MY_DISPLAY"))
	assert.True(strings.HasPrefix(materialized, "env:DISPLAY?fingerprint=v2-"), materialized)

	env, err := stemRunEnvRequirements(workspacePath)
	assert.Nil(err)
	assert.Equal(":99", env["MY_DISPLAY"])
}

func TestRootBuildStage1_VerifiesExistingEnvRequirementFingerprint(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	rootPath := filepath.Join(gardenPath, "dyd", "roots", "root-01")
	workspacePath := t.TempDir()
	requirementsPath := filepath.Join(rootPath, "dyd", "requirements")
	t.Setenv("DISPLAY", ":99")
	err, fingerprint := rootRequirementEnvValueFingerprint(":99")
	assert.Nil(err)

	writeFileForTest(t, filepath.Join(requirementsPath, "display"), rootRequirementEnvTargetString("DISPLAY", fingerprint))
	writeFileForTest(t, filepath.Join(workspacePath, "dyd", ".keep"), "")

	err, results := rootBuild_stage1(
		task.SERIAL_CONTEXT,
		rootBuild_stage1_request{
			Roots: &SafeRootsReference{
				BasePath: filepath.Join(gardenPath, "dyd", "roots"),
				Garden:   &SafeGardenReference{BasePath: gardenPath},
			},
			RootPath:                 rootPath,
			WorkspacePath:            workspacePath,
			SelectedRequirementsPath: requirementsPath,
		},
	)
	assert.Nil(err)
	assert.Empty(results)
	assert.Equal(rootRequirementEnvTargetString("DISPLAY", fingerprint), readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "requirements", "DISPLAY")))

	badWorkspacePath := t.TempDir()
	writeFileForTest(t, filepath.Join(badWorkspacePath, "dyd", ".keep"), "")
	t.Setenv("DISPLAY", ":100")
	err, _ = rootBuild_stage1(
		task.SERIAL_CONTEXT,
		rootBuild_stage1_request{
			Roots: &SafeRootsReference{
				BasePath: filepath.Join(gardenPath, "dyd", "roots"),
				Garden:   &SafeGardenReference{BasePath: gardenPath},
			},
			RootPath:                 rootPath,
			WorkspacePath:            badWorkspacePath,
			SelectedRequirementsPath: requirementsPath,
		},
	)
	assert.NotNil(err)
	assert.Contains(err.Error(), "fingerprint mismatch")
}

func TestStemRunEnvRequirements_VerifiesFingerprint(t *testing.T) {
	assert := assert.New(t)

	stemPath := t.TempDir()
	t.Setenv("DISPLAY", ":1")
	err, fingerprint := rootRequirementEnvValueFingerprint(":1")
	assert.Nil(err)

	writeFileForTest(
		t,
		filepath.Join(stemPath, "dyd", "requirements", "display"),
		rootRequirementEnvTargetString("DISPLAY", fingerprint),
	)

	env, err := stemRunEnvRequirements(stemPath)
	assert.Nil(err)
	assert.Equal(":1", env["DISPLAY"])

	t.Setenv("DISPLAY", ":2")
	_, err = stemRunEnvRequirements(stemPath)
	assert.NotNil(err)
	assert.Contains(err.Error(), "fingerprint mismatch")
}

func TestStemRunEnvRequirements_RejectsDuplicateInjectedName(t *testing.T) {
	assert := assert.New(t)

	stemPath := t.TempDir()
	t.Setenv("DISPLAY", ":1")

	writeFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "my-display"), "env:DISPLAY")
	writeFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "my.display"), "env:DISPLAY")

	_, err := stemRunEnvRequirements(stemPath)
	assert.NotNil(err)
	assert.Contains(err.Error(), "duplicate env requirement injects MY_DISPLAY")
}

func TestRootBuildRequirementsPrepare_PreservesRuntimeEnvRequirements(t *testing.T) {
	assert := assert.New(t)

	stemPath := t.TempDir()
	writeFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "app-display"), "env:display")

	err := rootBuild_requirementsPrepare(stemPath)
	assert.Nil(err)
	assert.Equal("env:display", readTrimmedFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "APP_DISPLAY")))
}

func TestRootRequirementsAddEnv_PreservesFingerprint(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	requirementsPath := filepath.Join(rootPath, "dyd", "requirements")
	err, fingerprint := rootRequirementEnvValueFingerprint("value")
	assert.Nil(err)

	requirements := &SafeRootRequirementsReference{
		BasePath: requirementsPath,
		Root:     &SafeRootReference{BasePath: rootPath},
	}
	err, requirement := requirements.AddEnv(task.SERIAL_CONTEXT, RootRequirementsAddEnvRequest{
		Alias:  "display",
		Target: rootRequirementEnvTargetString("DISPLAY", fingerprint),
	})
	assert.Nil(err)
	assert.NotNil(requirement)
	assert.Equal(rootRequirementEnvTargetString("DISPLAY", fingerprint), readTrimmedFileForTest(t, filepath.Join(requirementsPath, "display")))
}

func TestStemRunCommand_DoesNotMutateCallerEnv(t *testing.T) {
	assert := assert.New(t)

	gardenPath := t.TempDir()
	stemWithEnvPath := t.TempDir()
	stemWithoutEnvPath := t.TempDir()
	sharedEnv := map[string]string{"TERM": "dryad-test"}
	t.Setenv("DISPLAY", ":7")

	writeFileForTest(t, filepath.Join(stemWithEnvPath, "dyd", "commands", "dyd-stem-run"), "#!/usr/bin/env sh\n")
	writeFileForTest(t, filepath.Join(stemWithEnvPath, "dyd", "requirements", "display"), "env:DISPLAY")
	writeFileForTest(t, filepath.Join(stemWithoutEnvPath, "dyd", "commands", "dyd-stem-run"), "#!/usr/bin/env sh\n")

	withEnvInstance, err := StemRunCommand(StemRunRequest{
		Garden:   &SafeGardenReference{BasePath: gardenPath},
		StemPath: stemWithEnvPath,
		Env:      sharedEnv,
	})
	assert.Nil(err)
	if withEnvInstance != nil && withEnvInstance.Close != nil {
		defer withEnvInstance.Close()
	}
	assert.NotContains(sharedEnv, "DISPLAY")
	assert.Contains(withEnvInstance.Cmd.Env, "DISPLAY=:7")

	withoutEnvInstance, err := StemRunCommand(StemRunRequest{
		Garden:   &SafeGardenReference{BasePath: gardenPath},
		StemPath: stemWithoutEnvPath,
		Env:      sharedEnv,
	})
	assert.Nil(err)
	if withoutEnvInstance != nil && withoutEnvInstance.Close != nil {
		defer withoutEnvInstance.Close()
	}
	assert.NotContains(withoutEnvInstance.Cmd.Env, "DISPLAY=:7")
}

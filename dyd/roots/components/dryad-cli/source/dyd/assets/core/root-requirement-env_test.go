package core

import (
	"dryad/internal/filepath"
	"dryad/task"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootRequirementCanonicalEnvName(t *testing.T) {
	assert := assert.New(t)

	err, name := rootRequirementCanonicalEnvName("my-display.name")
	assert.Nil(err)
	assert.Equal("MY_DISPLAY_NAME", name)

	err, _ = rootRequirementCanonicalEnvName("1display")
	assert.NotNil(err)
}

func TestRootRequirementEnvDefaultNormalize(t *testing.T) {
	for _, tc := range []struct {
		name       string
		target     string
		normalized string
		defaultVal *string
	}{
		{name: "absent", target: "env:example", normalized: "env:EXAMPLE"},
		{name: "empty", target: "env:example?default=", normalized: "env:EXAMPLE?default=", defaultVal: stringPointerForEnvTest("")},
		{name: "encoded", target: "env:example?default=%20hello%26%2B%3D%3F%23%20", normalized: "env:EXAMPLE?default=+hello%26%2B%3D%3F%23+", defaultVal: stringPointerForEnvTest(" hello&+=?# ")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, normalized := RootRequirementEnvTargetNormalize(tc.target)
			require.NoError(t, err)
			assert.Equal(t, tc.normalized, normalized)
			err, spec, isEnv := rootRequirementParseEnvTarget(normalized)
			require.NoError(t, err)
			assert.True(t, isEnv)
			assert.Equal(t, tc.defaultVal, spec.Default)
		})
	}

	for _, target := range []string{
		"env:EXAMPLE?default=one&default=two",
		"env:EXAMPLE?default=%zz",
		"env:EXAMPLE?default=one;two",
		"env:EXAMPLE?default=one&unknown=two",
	} {
		t.Run(target, func(t *testing.T) {
			err, _ := RootRequirementEnvTargetNormalize(target)
			require.Error(t, err)
		})
	}
}

func stringPointerForEnvTest(value string) *string {
	return &value
}

func TestRootBuildStage1_EnvDefaults(t *testing.T) {
	const envName = "DRYAD_TEST_ENV_DEFAULT"
	for _, tc := range []struct {
		name    string
		query   string
		host    *string
		want    string
		missing bool
	}{
		{name: "fallback", query: "?default=0", want: "0"},
		{name: "host", query: "?default=0", host: stringPointerForEnvTest("123"), want: "123"},
		{name: "host equals default", query: "?default=0", host: stringPointerForEnvTest("0"), want: "0"},
		{name: "empty host", query: "?default=0", host: stringPointerForEnvTest(""), want: ""},
		{name: "empty default", query: "?default=", want: ""},
		{name: "encoded default", query: "?default=hello%20world%26%2B", want: "hello world&+"},
		{name: "required", missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(envName, "")
			if tc.host == nil {
				require.NoError(t, os.Unsetenv(envName))
			} else {
				t.Setenv(envName, *tc.host)
			}

			gardenPath := t.TempDir()
			rootPath := filepath.Join(gardenPath, "dyd", "roots", "root-01")
			requirementsPath := filepath.Join(rootPath, "dyd", "requirements")
			workspacePath := t.TempDir()
			writeFileForTest(t, filepath.Join(requirementsPath, "value"), "env:"+envName+tc.query)
			writeFileForTest(t, filepath.Join(workspacePath, "dyd", ".keep"), "")
			runtimeEnv, runtimeErr := stemRunEnvRequirements(rootPath)
			if tc.missing {
				require.Error(t, runtimeErr)
			} else {
				require.NoError(t, runtimeErr)
				assert.Equal(t, tc.want, runtimeEnv["VALUE"])
			}
			err, _ := rootBuild_stage1(task.SERIAL_CONTEXT, rootBuild_stage1_request{
				Roots: &SafeRootsReference{
					BasePath: filepath.Join(gardenPath, "dyd", "roots"),
					Garden:   &SafeGardenReference{BasePath: gardenPath},
				},
				RootPath:                 rootPath,
				WorkspacePath:            workspacePath,
				SelectedRequirementsPath: requirementsPath,
			})
			if tc.missing {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "is not set")
				return
			}
			require.NoError(t, err)
			materialized := readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "requirements", "VALUE"))
			err, spec, _ := rootRequirementParseEnvTarget(materialized)
			require.NoError(t, err)
			err, fingerprint := rootRequirementEnvValueFingerprint(tc.want)
			require.NoError(t, err)
			assert.Equal(t, fingerprint, spec.Fingerprint)
			err, originalSpec, _ := rootRequirementParseEnvTarget("env:" + envName + tc.query)
			require.NoError(t, err)
			assert.Equal(t, originalSpec.Default, spec.Default)

			env, err := stemRunEnvRequirements(workspacePath)
			require.NoError(t, err)
			assert.Equal(t, tc.want, env["VALUE"])
			t.Setenv(envName, "different")
			_, err = stemRunEnvRequirements(workspacePath)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "fingerprint mismatch")
		})
	}
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
	defaultValue := ":99"

	writeFileForTest(t, filepath.Join(requirementsPath, "display"), rootRequirementEnvTargetString("DISPLAY", fingerprint, &defaultValue))
	writeFileForTest(t, filepath.Join(workspacePath, "dyd", ".keep"), "")
	require.NoError(t, os.Unsetenv("DISPLAY"))

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
	assert.Equal(rootRequirementEnvTargetString("DISPLAY", fingerprint, &defaultValue), readTrimmedFileForTest(t, filepath.Join(workspacePath, "dyd", "requirements", "DISPLAY")))

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
		rootRequirementEnvTargetString("DISPLAY", fingerprint, nil),
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
	writeFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "app-display"), "env:display?default=%3A0")
	writeFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "empty"), "env:example?default=")

	err := rootBuild_requirementsPrepare(stemPath)
	assert.Nil(err)
	assert.Equal("env:display?default=%3A0", readTrimmedFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "APP_DISPLAY")))
	assert.Equal("env:example?default=", readTrimmedFileForTest(t, filepath.Join(stemPath, "dyd", "requirements", "EMPTY")))
}

func TestRootRequirementsAddEnv_PreservesDefaultAndFingerprint(t *testing.T) {
	assert := assert.New(t)

	rootPath := t.TempDir()
	requirementsPath := filepath.Join(rootPath, "dyd", "requirements")
	err, fingerprint := rootRequirementEnvValueFingerprint("value")
	assert.Nil(err)
	defaultValue := "fallback & value"

	requirements := &SafeRootRequirementsReference{
		BasePath: requirementsPath,
		Root:     &SafeRootReference{BasePath: rootPath},
	}
	err, requirement := requirements.AddEnv(task.SERIAL_CONTEXT, RootRequirementsAddEnvRequest{
		Alias:  "display",
		Target: rootRequirementEnvTargetString("DISPLAY", fingerprint, &defaultValue),
	})
	assert.Nil(err)
	assert.NotNil(requirement)
	assert.Equal(rootRequirementEnvTargetString("DISPLAY", fingerprint, &defaultValue), readTrimmedFileForTest(t, filepath.Join(requirementsPath, "display")))
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

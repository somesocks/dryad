package core

import (
	"dryad/internal/exec"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"

	zerolog "github.com/rs/zerolog"
)

var invalidChars = regexp.MustCompile(`[<>:"/\\|?*]`)

func sanitizePathSegment(s string) string {
	return invalidChars.ReplaceAllString(s, "-")
}

func resolveCommandOnPath(command string, pathValue string) (string, error) {
	for _, dir := range strings.Split(pathValue, string(os.PathListSeparator)) {
		if dir == "" {
			continue
		}

		candidate := filepath.Join(dir, command)
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		if info.Mode()&0o111 == 0 {
			continue
		}

		return candidate, nil
	}

	return "", fmt.Errorf("executable file not found in PATH: %s", command)
}

type StemRunRequest struct {
	Garden       *SafeGardenReference
	StemPath     string
	WorkingPath  string
	MainOverride string
	Context      string
	Env          map[string]string
	Args         []string
	Stdin        InputConfig
	Stdout       OutputConfig
	Stderr       OutputConfig
	InheritEnv   bool
}

func stemRunReservedEnvName(name string) bool {
	switch name {
	case "PATH",
		"HOME",
		"DYD_CONTEXT",
		"DYD_STEM",
		"DYD_GARDEN",
		"DYD_CLI_BIN",
		"DYD_OS",
		"DYD_ARCH",
		"DYD_LOG_LEVEL",
		"DOCKER_HOST":
		return true
	default:
		return false
	}
}

func stemRunEnvRequirements(stemPath string) (map[string]string, error) {
	requirementEnv := map[string]string{}
	requirementsPath := filepath.Join(stemPath, "dyd", "requirements")
	requirementEntries, err := os.ReadDir(requirementsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return requirementEnv, nil
		}
		return nil, err
	}

	for _, requirementEntry := range requirementEntries {
		if requirementEntry.IsDir() {
			continue
		}

		requirementPath := filepath.Join(requirementsPath, requirementEntry.Name())
		requirementBytes, err := os.ReadFile(requirementPath)
		if err != nil {
			return nil, err
		}

		err, envSpec, isEnv := rootRequirementParseEnvTarget(string(requirementBytes))
		if err != nil {
			return nil, fmt.Errorf("invalid env requirement %s: %w", requirementEntry.Name(), err)
		}
		if !isEnv {
			continue
		}

		err, requirementName, condition := rootRequirementParseName(requirementEntry.Name())
		if err != nil {
			return nil, err
		}
		if len(condition) > 0 {
			return nil, fmt.Errorf("conditional env requirement is not valid in a stem: %s", requirementEntry.Name())
		}

		err, injectName := rootRequirementCanonicalEnvName(requirementName)
		if err != nil {
			return nil, err
		}
		if stemRunReservedEnvName(injectName) {
			return nil, fmt.Errorf("env requirement cannot inject reserved env var: %s", injectName)
		}

		envValue, exists := rootRequirementEnvValue(envSpec.Name, envSpec.Default)
		if !exists {
			return nil, fmt.Errorf("missing env requirement %s: host env %s is not set", injectName, envSpec.Name)
		}

		if envSpec.Fingerprint != "" {
			err, envFingerprint := rootRequirementEnvValueFingerprint(envValue)
			if err != nil {
				return nil, err
			}
			if envFingerprint != envSpec.Fingerprint {
				return nil, fmt.Errorf("env requirement %s fingerprint mismatch for host env %s", injectName, envSpec.Name)
			}
		}

		if _, exists := requirementEnv[injectName]; exists {
			return nil, fmt.Errorf("duplicate env requirement injects %s", injectName)
		}
		requirementEnv[injectName] = envValue
	}

	return requirementEnv, nil
}

func stemRun_prepContext(request StemRunRequest) (string, error) {
	var gardenPath string
	var err error
	var context string

	context = request.Context
	if context == "" {
		context = "default"
	}

	gardenPath = request.Garden.BasePath

	contextPath := filepath.Join(gardenPath, "dyd", "heap", "contexts", context)
	err = os.MkdirAll(contextPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	return contextPath, nil
}

type StemRunInstance struct {
	Cmd   *exec.Cmd
	Close func() error
}

func StemRunCommand(request StemRunRequest) (*StemRunInstance, error) {
	stdin, err := request.Stdin.Resolve(1, 1, false)
	if err != nil {
		return nil, err
	}
	if err := request.Stdout.validate(); err != nil {
		return nil, fmt.Errorf("stdout: %w", err)
	}
	if err := request.Stderr.validate(); err != nil {
		return nil, fmt.Errorf("stderr: %w", err)
	}
	var workingPath = request.WorkingPath
	var stemPath = request.StemPath
	var env = map[string]string{}
	var args = request.Args
	var gardenPath string
	var closers []io.Closer
	// Close partially opened log files if subsequent preparation fails.
	prepared := false
	defer func() {
		if !prepared {
			for _, closer := range closers {
				closer.Close()
			}
		}
	}()

	for key, value := range request.Env {
		env[key] = value
	}

	if !filepath.IsAbs(stemPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		stemPath = filepath.Join(cwd, stemPath)
	}

	gardenPath = request.Garden.BasePath

	contextPath, err := stemRun_prepContext(request)
	if err != nil {
		return nil, err
	}

	// prepare by getting the executable path
	dryadPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dryadPath, err = filepath.EvalSymlinks(dryadPath)
	if err != nil {
		return nil, err
	}
	dryadBin := dryadPath
	dryadPath = filepath.Dir(dryadPath)

	stemPathEnv := BuildPlatformPath(stemPath, dryadPath)

	requirementEnv, err := stemRunEnvRequirements(stemPath)
	if err != nil {
		return nil, err
	}
	for key, value := range requirementEnv {
		if existingValue, exists := env[key]; exists {
			if existingValue != value {
				return nil, fmt.Errorf("env requirement conflicts with explicit env value for %s", key)
			}
			continue
		}
		env[key] = value
	}

	var command string
	if request.MainOverride != "" {
		override := request.MainOverride
		if filepath.IsAbs(override) || strings.ContainsRune(override, os.PathSeparator) {
			command = override
		} else {
			resolved, err := resolveCommandOnPath(override, stemPathEnv)
			if err != nil {
				return nil, fmt.Errorf("missing stem main %q: %w", override, err)
			}
			command = resolved
		}
	} else {
		command = stemPath + "/dyd/commands/dyd-stem-run"
	}

	info, err := os.Stat(command)
	if err != nil {
		return nil, fmt.Errorf("missing stem main %q: %w", command, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("stem main is a directory %q", command)
	}
	if info.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("stem main is not executable %q", command)
	}

	cmd := exec.Command(
		command,
		args...,
	)

	cmd.Dir = workingPath

	// prepare env
	cmd.Env = []string{}

	if request.InheritEnv {
		cmd.Env = append(
			cmd.Env, os.Environ()...)
	}

	for key, val := range env {
		cmd.Env = append(cmd.Env, key+"="+val)
	}

	if stdin.Mode == InputInherit {
		cmd.Stdin = os.Stdin
	}

	// optionally pipe the exec logs to us
	if request.Stdout.Mode == OutputInherit {
		cmd.Stdout = os.Stdout
	} else if request.Stdout.Mode == OutputLog {
		var outputPath string

		if request.Stdout.Name != "" {
			outputPath = filepath.Join(request.Stdout.Path, request.Stdout.Name)
		} else {
			relStemPath, err := filepath.Rel(gardenPath, stemPath)
			if err != nil {
				return nil, err
			}

			logFile := "dyd-stem-run--" + sanitizePathSegment(relStemPath) + ".out"
			outputPath = filepath.Join(request.Stdout.Path, logFile)
		}

		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = file
		closers = append(closers, file)
	}

	// optionally pipe the exec stderr to us
	if request.Stderr.Mode == OutputInherit {
		cmd.Stderr = os.Stderr
	} else if request.Stderr.Mode == OutputLog {
		var outputPath string

		if request.Stderr.Name != "" {
			outputPath = filepath.Join(request.Stderr.Path, request.Stderr.Name)
		} else {
			relStemPath, err := filepath.Rel(gardenPath, stemPath)
			if err != nil {
				return nil, err
			}

			logFile := "dyd-stem-run--" + sanitizePathSegment(relStemPath) + ".err"
			outputPath = filepath.Join(request.Stderr.Path, logFile)
		}

		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		cmd.Stderr = file
		closers = append(closers, file)
	}

	envPath := "PATH=" + stemPathEnv
	cmd.Env = append(
		cmd.Env,
		envPath,
		"HOME="+contextPath,
		"DYD_CONTEXT="+contextPath,
		"DYD_STEM="+stemPath,
		"DYD_GARDEN="+gardenPath,
		"DYD_CLI_BIN="+dryadBin,
		"DYD_OS="+runtime.GOOS,
		"DYD_ARCH="+runtime.GOARCH,
		"DYD_LOG_LEVEL="+zerolog.GlobalLevel().String(),
	)

	// Always override DOCKER_HOST because context HOME changes can invalidate user paths.
	if dockerSock := GetDockerSockPath(); dockerSock != "" {
		cmd.Env = append(cmd.Env, "DOCKER_HOST=unix://"+dockerSock)
	}

	instance := &StemRunInstance{
		Cmd: cmd,
		Close: func() error {
			var firstErr error
			for _, closer := range closers {
				if err := closer.Close(); err != nil && firstErr == nil {
					firstErr = err
				}
			}
			return firstErr
		},
	}

	prepared = true
	return instance, nil
}

func StemRun(request StemRunRequest) error {
	instance, err := StemRunCommand(request)
	if err != nil {
		return err
	}
	if instance.Close != nil {
		defer instance.Close()
	}

	err = instance.Cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

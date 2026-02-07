package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	Garden *SafeGardenReference
	StemPath     string
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
	JoinStderr   bool
	LogStderr    struct {
			Path string
			Name string
		}
	InheritEnv   bool
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
	var workingPath = request.WorkingPath
	var stemPath = request.StemPath
	var env = request.Env
	var args = request.Args
	var gardenPath string
	var err error
	var closers []io.Closer

	if env == nil {
		env = make(map[string]string)
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

	cmd.Stdin = os.Stdin

	// optionally pipe the exec logs to us
	if request.JoinStdout {
		cmd.Stdout = os.Stdout
	} else if request.LogStdout.Path != "" {
		var outputPath string

		if request.LogStdout.Name != "" {
			outputPath = filepath.Join(request.LogStdout.Path, request.LogStdout.Name)				
		} else {
			relStemPath, err := filepath.Rel(gardenPath, stemPath)
			if err != nil {
				return nil, err
			}
	
			logFile := "dyd-stem-run--" + sanitizePathSegment(relStemPath) + ".out"
			outputPath = filepath.Join(request.LogStdout.Path, logFile)	
		}

		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = file
		closers = append(closers, file)
	}

	// optionally pipe the exec stderr to us
	if request.JoinStderr {
		cmd.Stderr = os.Stderr
	} else if request.LogStderr.Path != "" {
		var outputPath string

		if request.LogStderr.Name != "" {
			outputPath = filepath.Join(request.LogStderr.Path, request.LogStderr.Name)	
		} else {
			relStemPath, err := filepath.Rel(gardenPath, stemPath)
			if err != nil {
				return nil, err
			}
		
			logFile := "dyd-stem-run--" + sanitizePathSegment(relStemPath) + ".err"
			outputPath = filepath.Join(request.LogStderr.Path, logFile)	
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

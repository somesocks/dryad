package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// BuildPlatformPath constructs PATH with platform-appropriate paths.
// On macOS, this includes Homebrew paths for Apple Silicon compatibility.
func BuildPlatformPath(stemPath, dryadPath string) string {
	basePath := fmt.Sprintf(
		"%s/dyd/commands:%s/dyd/path:%s:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		stemPath, stemPath, dryadPath,
	)

	if runtime.GOOS == "darwin" {
		// Apple Silicon uses /opt/homebrew, Intel uses /usr/local (already included)
		return basePath + ":/opt/homebrew/bin:/opt/homebrew/sbin"
	}
	return basePath
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func homeDockerSock() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(homeDir, ".docker", "run", "docker.sock")
	if pathExists(path) {
		return path
	}
	return ""
}

// GetDockerSockPath returns the platform-appropriate Docker socket path.
// Returns empty string if no socket path is found.
func GetDockerSockPath() string {
	if runtime.GOOS == "darwin" {
		if sock := homeDockerSock(); sock != "" {
			return sock
		}
		if pathExists("/var/run/docker.sock") {
			return "/var/run/docker.sock"
		}
		return ""
	}

	if runtime.GOOS == "linux" {
		if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
			path := filepath.Join(xdg, "docker.sock")
			if pathExists(path) {
				return path
			}
		}
		if sock := homeDockerSock(); sock != "" {
			return sock
		}
	}

	if pathExists("/var/run/docker.sock") {
		return "/var/run/docker.sock"
	}

	return ""
}

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

// GetDockerSockPath returns the platform-appropriate Docker socket path.
// On macOS with Docker Desktop, the socket is at $HOME/.docker/run/docker.sock.
// On Linux (or if macOS socket doesn't exist), returns /var/run/docker.sock.
func GetDockerSockPath() string {
	if runtime.GOOS == "darwin" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			macOSSock := filepath.Join(homeDir, ".docker", "run", "docker.sock")
			if _, err := os.Stat(macOSSock); err == nil {
				return macOSSock
			}
		}
	}
	return "/var/run/docker.sock"
}

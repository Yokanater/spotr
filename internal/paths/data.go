package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const dataDirEnv = "SPOTR_DATA_DIR"

// DatabasePath returns the spotr database path, creating its parent directory.
// An explicit directory takes precedence over SPOTR_DATA_DIR and OS defaults.
func DatabasePath(explicitDir string) (string, error) {
	dir := explicitDir
	if dir == "" {
		dir = os.Getenv(dataDirEnv)
	}
	if dir == "" {
		var err error
		dir, err = defaultDataDir()
		if err != nil {
			return "", err
		}
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return filepath.Join(dir, "spotr.db"), nil
}

func defaultDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "spotr"), nil
	case "windows":
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "spotr"), nil
		}
		return filepath.Join(home, "AppData", "Local", "spotr"), nil
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "spotr"), nil
		}
		return filepath.Join(home, ".local", "share", "spotr"), nil
	}
}

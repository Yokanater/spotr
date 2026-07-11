package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatabasePathExplicitDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "data")
	path, err := DatabasePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "spotr.db"); path != want {
		t.Fatalf("DatabasePath() = %q, want %q", path, want)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("data directory was not created: %v", err)
	}
}

func TestDatabasePathEnvironmentOverride(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "env-data")
	t.Setenv(dataDirEnv, dir)

	path, err := DatabasePath("")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "spotr.db"); path != want {
		t.Fatalf("DatabasePath() = %q, want %q", path, want)
	}
}

func TestExplicitDirectoryWinsOverEnvironment(t *testing.T) {
	explicit := filepath.Join(t.TempDir(), "explicit")
	t.Setenv(dataDirEnv, filepath.Join(t.TempDir(), "environment"))

	path, err := DatabasePath(explicit)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(explicit, "spotr.db"); path != want {
		t.Fatalf("DatabasePath() = %q, want %q", path, want)
	}
}

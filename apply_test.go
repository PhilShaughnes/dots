package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyDotfile(t *testing.T) {
	t.Run("creates symlink when missing", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "repo", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.WriteFile(src, []byte("# zshrc"), 0644)

		s := entryStatus{name: "zshrc", kind: "file", state: stateMissing, src: src, dest: dest}
		if err := applyDotfile(s); err != nil {
			t.Fatalf("applyDotfile() error = %v", err)
		}

		target, err := os.Readlink(dest)
		if err != nil {
			t.Fatalf("dest is not a symlink: %v", err)
		}
		if target != src {
			t.Errorf("symlink target = %q, want %q", target, src)
		}
	})

	t.Run("skips when already ok", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "repo", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.MkdirAll(filepath.Dir(dest), 0755)
		os.WriteFile(src, []byte("# zshrc"), 0644)
		os.Symlink(src, dest)

		s := entryStatus{name: "zshrc", kind: "file", state: stateOK, src: src, dest: dest}
		if err := applyDotfile(s); err != nil {
			t.Fatalf("applyDotfile() error = %v", err)
		}
		// dest should still be the original symlink, unchanged
		target, _ := os.Readlink(dest)
		if target != src {
			t.Errorf("symlink target changed: got %q, want %q", target, src)
		}
	})

	t.Run("skips conflict without error", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "repo", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.MkdirAll(filepath.Dir(dest), 0755)
		os.WriteFile(src, []byte("# src"), 0644)
		os.WriteFile(dest, []byte("# existing"), 0644)

		s := entryStatus{name: "zshrc", kind: "file", state: stateConflict, src: src, dest: dest}
		if err := applyDotfile(s); err != nil {
			t.Fatalf("applyDotfile() should not error on conflict, got: %v", err)
		}
		// dest should still be a real file
		info, _ := os.Lstat(dest)
		if info.Mode()&os.ModeSymlink != 0 {
			t.Error("conflict file was replaced with symlink — should have been skipped")
		}
	})

	t.Run("creates parent dirs", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "repo", ".config", "jj", "config.toml")
		dest := filepath.Join(dir, "home", ".config", "jj", "config.toml")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.WriteFile(src, []byte("[user]"), 0644)

		s := entryStatus{name: "jj", kind: "file", state: stateMissing, src: src, dest: dest}
		if err := applyDotfile(s); err != nil {
			t.Fatalf("applyDotfile() error = %v", err)
		}
		if _, err := os.Lstat(dest); err != nil {
			t.Errorf("dest not created: %v", err)
		}
	})
}

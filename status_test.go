package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDotfileState(t *testing.T) {
	t.Run("ok — correct symlink", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.MkdirAll(filepath.Dir(dest), 0755)
		os.WriteFile(src, []byte("# zshrc"), 0644)
		os.Symlink(src, dest)

		if got := dotfileState(src, dest); got != stateOK {
			t.Errorf("want stateOK, got %s", got)
		}
	})

	t.Run("missing — nothing at dest", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.WriteFile(src, []byte("# zshrc"), 0644)

		if got := dotfileState(src, dest); got != stateMissing {
			t.Errorf("want stateMissing, got %s", got)
		}
	})

	t.Run("conflict — real file at dest", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.MkdirAll(filepath.Dir(dest), 0755)
		os.WriteFile(src, []byte("# src"), 0644)
		os.WriteFile(dest, []byte("# existing"), 0644)

		if got := dotfileState(src, dest); got != stateConflict {
			t.Errorf("want stateConflict, got %s", got)
		}
	})

	t.Run("wrong-target — symlink points elsewhere", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src", ".zshrc")
		other := filepath.Join(dir, "other", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		os.MkdirAll(filepath.Dir(src), 0755)
		os.MkdirAll(filepath.Dir(other), 0755)
		os.MkdirAll(filepath.Dir(dest), 0755)
		os.WriteFile(src, []byte("# src"), 0644)
		os.WriteFile(other, []byte("# other"), 0644)
		os.Symlink(other, dest)

		if got := dotfileState(src, dest); got != stateWrongTarget {
			t.Errorf("want stateWrongTarget, got %s", got)
		}
	})

	t.Run("src-missing — src does not exist", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src", ".zshrc")
		dest := filepath.Join(dir, "home", ".zshrc")

		if got := dotfileState(src, dest); got != stateSrcMissing {
			t.Errorf("want stateSrcMissing, got %s", got)
		}
	})
}

func TestRepoState(t *testing.T) {
	t.Run("ok — git repo exists", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "nvim")
		os.MkdirAll(filepath.Join(dest, ".git"), 0755)

		if got := repoState(dest); got != stateCloned {
			t.Errorf("want stateCloned, got %s", got)
		}
	})

	t.Run("ok — jj repo exists", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "nvim")
		os.MkdirAll(filepath.Join(dest, ".jj"), 0755)

		if got := repoState(dest); got != stateCloned {
			t.Errorf("want stateCloned, got %s", got)
		}
	})

	t.Run("not-cloned — dest missing", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "nvim")

		if got := repoState(dest); got != stateNotCloned {
			t.Errorf("want stateNotCloned, got %s", got)
		}
	})

	t.Run("conflict — dir exists but not a repo", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "nvim")
		os.MkdirAll(dest, 0755)

		if got := repoState(dest); got != stateConflict {
			t.Errorf("want stateConflict, got %s", got)
		}
	})
}

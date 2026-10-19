package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestApplyDotfile(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(src, dest string)
		state     entryState
		wantState entryState
		wantLink  bool // dest should be a symlink after apply
	}{
		{
			name: "empty — creates symlink",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.WriteFile(src, []byte("# zshrc"), 0644)
			},
			state:     stateEmpty,
			wantState: stateChanged,
			wantLink:  true,
		},
		{
			name: "ok — already correct",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.MkdirAll(filepath.Dir(dest), 0755)
				os.WriteFile(src, []byte("# zshrc"), 0644)
				os.Symlink(src, dest)
			},
			state:     stateOK,
			wantState: stateOK,
			wantLink:  true,
		},
		{
			name: "blocked — skipped without error",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.MkdirAll(filepath.Dir(dest), 0755)
				os.WriteFile(src, []byte("# src"), 0644)
				os.WriteFile(dest, []byte("# existing"), 0644)
			},
			state:     stateBlocked,
			wantState: stateBlocked,
			wantLink:  false,
		},
		{
			name: "empty — creates parent dirs",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.WriteFile(src, []byte("[user]"), 0644)
			},
			state:     stateEmpty,
			wantState: stateChanged,
			wantLink:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "repo", ".zshrc")
			dest := filepath.Join(dir, "home", ".zshrc")
			tt.setup(src, dest)

			s := entryStatus{name: "zshrc", kind: "dotfile", state: tt.state, src: src, dest: dest}
			newState, err := applyDotfile(s)
			if err != nil {
				t.Fatalf("applyDotfile() unexpected error: %v", err)
			}
			if newState != tt.wantState {
				t.Errorf("state = %s, want %s", newState, tt.wantState)
			}
			info, _ := os.Lstat(dest)
			isLink := info != nil && info.Mode()&os.ModeSymlink != 0
			if isLink != tt.wantLink {
				t.Errorf("dest is symlink = %v, want %v", isLink, tt.wantLink)
			}
		})
	}
}

func TestApplyRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	tests := []struct {
		name  string
		setup func(dest string)
	}{
		{
			name:  "empty — clones into missing dest",
			setup: func(dest string) {},
		},
		{
			name: "empty — clones into existing empty dir",
			setup: func(dest string) {
				os.MkdirAll(dest, 0755)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			url := filepath.Join(dir, "remote.git")
			if out, err := exec.Command("git", "init", "-q", "--bare", url).CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, out)
			}
			dest := filepath.Join(dir, "home", "repo")
			tt.setup(dest)

			s := entryStatus{name: "repo", kind: "repo", state: repoState(dest), url: url, dest: dest}
			if s.state != stateEmpty {
				t.Fatalf("repoState = %s, want empty", s.state)
			}
			newState, err := applyRepo(s)
			if err != nil {
				t.Fatalf("applyRepo() unexpected error: %v", err)
			}
			if newState != stateChanged {
				t.Errorf("state = %s, want %s", newState, stateChanged)
			}
			if got := repoState(dest); got != stateOK {
				t.Errorf("repoState after apply = %s, want ok", got)
			}
		})
	}
}

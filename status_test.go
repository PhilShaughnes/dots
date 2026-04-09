package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDotfileState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(src, dest string)
		want  entryState
	}{
		{
			name: "ok — correct symlink",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.MkdirAll(filepath.Dir(dest), 0755)
				os.WriteFile(src, []byte("# zshrc"), 0644)
				os.Symlink(src, dest)
			},
			want: stateOK,
		},
		{
			name: "empty — nothing at dest",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.WriteFile(src, []byte("# zshrc"), 0644)
			},
			want: stateEmpty,
		},
		{
			name: "blocked — real file at dest",
			setup: func(src, dest string) {
				os.MkdirAll(filepath.Dir(src), 0755)
				os.MkdirAll(filepath.Dir(dest), 0755)
				os.WriteFile(src, []byte("# src"), 0644)
				os.WriteFile(dest, []byte("# existing"), 0644)
			},
			want: stateBlocked,
		},
		{
			name: "blocked — symlink points elsewhere",
			setup: func(src, dest string) {
				dir := filepath.Dir(src)
				other := filepath.Join(dir, "other")
				os.MkdirAll(dir, 0755)
				os.MkdirAll(filepath.Dir(dest), 0755)
				os.WriteFile(src, []byte("# src"), 0644)
				os.WriteFile(other, []byte("# other"), 0644)
				os.Symlink(other, dest)
			},
			want: stateBlocked,
		},
		{
			name:  "src-missing — src does not exist",
			setup: func(src, dest string) {},
			want:  stateSrcMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "src", ".zshrc")
			dest := filepath.Join(dir, "home", ".zshrc")
			tt.setup(src, dest)
			if got := dotfileState(src, dest); got != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

func TestRepoState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(dest string)
		want  entryState
	}{
		{
			name: "ok — git repo exists",
			setup: func(dest string) {
				os.MkdirAll(filepath.Join(dest, ".git"), 0755)
			},
			want: stateOK,
		},
		{
			name: "ok — jj repo exists",
			setup: func(dest string) {
				os.MkdirAll(filepath.Join(dest, ".jj"), 0755)
			},
			want: stateOK,
		},
		{
			name:  "empty — dest does not exist",
			setup: func(dest string) {},
			want:  stateEmpty,
		},
		{
			name: "blocked — dir exists but not a repo",
			setup: func(dest string) {
				os.MkdirAll(dest, 0755)
			},
			want: stateBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			dest := filepath.Join(dir, "nvim")
			tt.setup(dest)
			if got := repoState(dest); got != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

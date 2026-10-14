package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CLI tests run the built binary, since stdin handling only shows up at the
// process boundary.

var dotsBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "dots-cli")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dotsBin = filepath.Join(dir, "dots")
	if out, err := exec.Command("go", "build", "-o", dotsBin, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build dots: %v\n%s", err, out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// run executes dots with the given stdin and returns stdout.
// Fails the test if dots doesn't exit within 5s.
func run(t *testing.T, stdin io.Reader, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, dotsBin, args...)
	cmd.Stdin = stdin
	cmd.Env = append(os.Environ(), "DOTS_ROOT=", "DOTS_SRCROOT=")
	out, _ := cmd.Output() // list exits 1 when entries aren't ok; only output matters here
	if ctx.Err() != nil {
		t.Fatalf("dots %v: timed out (blocked reading stdin?)", args)
	}
	return string(out)
}

// writeManifest writes a manifest with dotfiles "a" and "b" (srcs present,
// dests empty) and returns the manifest path and its directory.
func writeManifest(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	var b strings.Builder
	for _, name := range []string{"a", "b"} {
		src := filepath.Join(dir, "src", name)
		os.MkdirAll(filepath.Dir(src), 0755)
		os.WriteFile(src, []byte(name), 0644)
		fmt.Fprintf(&b, "[[dotfiles]]\nname = %q\nsrc = %q\ndest = %q\n\n", name, src, filepath.Join(dir, name))
	}
	path := filepath.Join(dir, "m.toml")
	os.WriteFile(path, []byte(b.String()), 0644)
	return path, dir
}

func TestCLIListNames(t *testing.T) {
	manifest, _ := writeManifest(t)

	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"piped stdin ignored without -n -", "a\n", nil, "a\nb\n"},
		{"-n - reads names", "b\n", []string{"-n", "-"}, "b\n"},
		{"-n - with empty stdin matches nothing", "", []string{"-n", "-"}, ""},
		{"-n - combines with -n", "a\n", []string{"-n", "-", "-n", "b"}, "a\nb\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"list", "-f", manifest, "-o", "name"}, tt.args...)
			if got := run(t, strings.NewReader(tt.stdin), args...); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// An agent harness or CI job may hand dots a pipe that never closes.
func TestCLIOpenStdinDoesNotBlock(t *testing.T) {
	manifest, _ := writeManifest(t)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if got := run(t, r, "list", "-f", manifest, "-o", "name"); got != "a\nb\n" {
		t.Errorf("got %q, want %q", got, "a\nb\n")
	}
}

// The dangerous case: an empty pipe used to mean "no filter", so apply
// would touch every entry.
func TestCLIApplyEmptyStdinAppliesNothing(t *testing.T) {
	manifest, dir := writeManifest(t)

	run(t, strings.NewReader(""), "apply", "-f", manifest, "-n", "-")

	for _, name := range []string{"a", "b"} {
		if _, err := os.Lstat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("%s: expected no symlink, got err=%v", name, err)
		}
	}
}

func TestCLIAddStdin(t *testing.T) {
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  []string // dests expected in manifest; nil means manifest not created
	}{
		{"- reads dests", "~/.zshrc\n~/.vimrc\n", []string{"-"}, []string{"~/.zshrc", "~/.vimrc"}},
		{"- mixes with args", "~/.vimrc\n", []string{"~/.zshrc", "-"}, []string{"~/.zshrc", "~/.vimrc"}},
		{"- with empty stdin adds nothing", "", []string{"-"}, nil},
		{"piped stdin ignored without -", "~/.vimrc\n", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := filepath.Join(t.TempDir(), "m.toml")
			args := append([]string{"add", "-f", manifest}, tt.args...)
			run(t, strings.NewReader(tt.stdin), args...)

			data, err := os.ReadFile(manifest)
			if tt.want == nil {
				if !os.IsNotExist(err) {
					t.Fatalf("expected no manifest, got %q (err=%v)", data, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, dest := range tt.want {
				if !strings.Contains(string(data), fmt.Sprintf("dest = %q", dest)) {
					t.Errorf("manifest missing dest %s:\n%s", dest, data)
				}
			}
		})
	}
}

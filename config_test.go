package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		toml    string
		wantErr bool
	}{
		{
			name: "valid config",
			toml: `
[[dotfiles]]
name = "zshrc"
src  = "~/dotfiles/.zshrc"
dest = "~/.zshrc"

[[repos]]
name = "nvim"
url  = "git@github.com:user/nvim.git"
dest = "~/.config/nvim"
`,
		},
		{
			name: "duplicate name across sections",
			toml: `
[[dotfiles]]
name = "nvim"
src  = "~/dotfiles/.zshrc"
dest = "~/.zshrc"

[[repos]]
name = "nvim"
url  = "git@github.com:user/nvim.git"
dest = "~/.config/nvim"
`,
			wantErr: true,
		},
		{
			name: "duplicate name within dotfiles",
			toml: `
[[dotfiles]]
name = "zshrc"
src  = "~/dotfiles/.zshrc"
dest = "~/.zshrc"

[[dotfiles]]
name = "zshrc"
src  = "~/dotfiles/.zshrc2"
dest = "~/.zshrc2"
`,
			wantErr: true,
		},
		{
			name: "missing dotfile name",
			toml: `
[[dotfiles]]
src  = "~/dotfiles/.zshrc"
dest = "~/.zshrc"
`,
			wantErr: true,
		},
		{
			name: "missing repo name",
			toml: `
[[repos]]
url  = "git@github.com:user/nvim.git"
dest = "~/.config/nvim"
`,
			wantErr: true,
		},
		{
			name: "empty config is valid",
			toml: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "dots-*.toml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())
			f.WriteString(tt.toml)
			f.Close()

			_, err = loadConfigs([]string{f.Name()})
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo", filepath.Join(home, "foo")},
		{"~/.config/nvim", filepath.Join(home, ".config/nvim")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

package main

import (
	"os"
	"testing"
)

func TestInferName(t *testing.T) {
	tests := []struct {
		dest string
		want string
	}{
		{"~/.zshrc", "zshrc"},
		{"~/.tmux.conf", "tmux"},
		{"~/.config/nvim/init.lua", "init"},
		{"~/.gitconfig", "gitconfig"},
		{"~/dotfiles/env.base", "env"},
		{"~/.ssh/config", "config"},
		{"~/.config/alacritty/alacritty.toml", "alacritty"},
		// basename with no dot prefix and no extension
		{"~/scripts/myscript", "myscript"},
	}
	for _, tt := range tests {
		t.Run(tt.dest, func(t *testing.T) {
			got := inferName(tt.dest)
			if got != tt.want {
				t.Errorf("inferName(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestMirrorPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		dest string
		root string
		want string
	}{
		{"~/.zshrc", "~/dotfiles", "~/dotfiles/.zshrc"},
		{"~/.tmux.conf", "~/dotfiles/", "~/dotfiles/.tmux.conf"},
		{"~/.config/nvim/init.lua", "~/dotfiles", "~/dotfiles/.config/nvim/init.lua"},
		// absolute dest path
		{home + "/.zshrc", "~/dotfiles", "~/dotfiles/.zshrc"},
	}
	for _, tt := range tests {
		t.Run(tt.dest, func(t *testing.T) {
			got := mirrorPath(tt.dest, tt.root)
			if got != tt.want {
				t.Errorf("mirrorPath(%q, %q) = %q, want %q", tt.dest, tt.root, got, tt.want)
			}
		})
	}
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Dotfiles []DotfileEntry `toml:"dotfiles"`
	Repos    []RepoEntry    `toml:"repos"`
}

type DotfileEntry struct {
	Name string `toml:"name"`
	Src  string `toml:"src"`
	Dest string `toml:"dest"`
}

type RepoEntry struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
	Dest string `toml:"dest"`
}

func loadConfigs(paths []string) (*Config, error) {
	var merged Config
	for _, p := range paths {
		var cfg Config
		if _, err := toml.DecodeFile(expandPath(p), &cfg); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		merged.Dotfiles = append(merged.Dotfiles, cfg.Dotfiles...)
		merged.Repos = append(merged.Repos, cfg.Repos...)
	}
	if err := validateConfig(&merged); err != nil {
		return nil, err
	}
	return &merged, nil
}

func validateConfig(cfg *Config) error {
	seen := map[string]bool{}
	for _, d := range cfg.Dotfiles {
		if d.Name == "" {
			return fmt.Errorf("dotfile entry missing name (src: %s)", d.Src)
		}
		if seen[d.Name] {
			return fmt.Errorf("duplicate entry name: %s", d.Name)
		}
		seen[d.Name] = true
	}
	for _, r := range cfg.Repos {
		if r.Name == "" {
			return fmt.Errorf("repo entry missing name (url: %s)", r.URL)
		}
		if seen[r.Name] {
			return fmt.Errorf("duplicate entry name: %s", r.Name)
		}
		seen[r.Name] = true
	}
	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

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

// loadManifests resolves default manifests if needed and loads configs,
// printing an error and returning false on failure.
func loadManifests(o *opts) (*Config, bool) {
	if len(o.files) == 0 {
		files, err := defaultManifests()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return nil, false
		}
		o.files = files
	}
	cfg, err := loadConfigs(o.files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return nil, false
	}
	return cfg, true
}

func loadConfigs(paths []string) (*Config, error) {
	var merged Config
	for _, p := range paths {
		var cfg Config
		if _, err := toml.DecodeFile(expandPath(resolveManifestPath(p)), &cfg); err != nil {
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

// resolveManifestPath resolves a manifest file path.
// Relative paths are resolved against DOTS_ROOT if set, otherwise cwd.
func resolveManifestPath(path string) string {
	if strings.HasPrefix(path, "~/") || filepath.IsAbs(path) {
		return path
	}
	if root := os.Getenv("DOTS_ROOT"); root != "" {
		return filepath.Join(expandPath(root), path)
	}
	return path
}

// defaultManifests returns the default manifest paths when no -f is given.
// Uses $DOTS_ROOT/dots.toml if DOTS_ROOT is set, otherwise dots.toml in cwd.
func defaultManifests() ([]string, error) {
	path := resolveManifestPath("dots.toml")
	if _, err := os.Stat(expandPath(path)); os.IsNotExist(err) {
		return nil, fmt.Errorf("no manifest found (tried %s) — use -f or set DOTS_ROOT", path)
	}
	return []string{path}, nil
}

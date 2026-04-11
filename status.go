package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type entryState int

const (
	stateOK        entryState = iota // symlink correct (dotfile) or repo present
	stateChanged                     // just fixed by apply; displays as ok*
	stateEmpty                       // nothing at dest
	stateBlocked                     // dest exists but wrong; needs manual intervention
	stateSrcMissing                  // src missing in dotfiles repo (dotfile only)
)

func (s entryState) String() string {
	switch s {
	case stateOK:
		return "ok"
	case stateChanged:
		return "ok*"
	case stateEmpty:
		return "empty"
	case stateBlocked:
		return "blocked"
	case stateSrcMissing:
		return "src-missing"
	}
	return "unknown"
}

type entryStatus struct {
	name  string
	kind  string // "dotfile" or "repo"
	state entryState
	src   string // dotfiles: local path in repo
	url   string // repos: git remote URL
	dest  string
}

func cmdList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	var o opts
	registerFlags(fs, &o)
	fs.Parse(args)

	if len(o.files) == 0 {
		fmt.Fprintln(os.Stderr, "error: -f required")
		fs.Usage()
		return 1
	}

	cfg, err := loadConfigs(o.files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	names := collectNames(o.names)
	exit := 0
	for _, s := range collectStatuses(cfg, names) {
		if !matchType(s, o.types) || !matchState(s.state, o.states) {
			continue
		}
		outputEntry(s, o.fields)
		if s.state != stateOK {
			exit = 1
		}
	}
	return exit
}

func collectStatuses(cfg *Config, names map[string]bool) []entryStatus {
	var out []entryStatus
	for _, d := range cfg.Dotfiles {
		if len(names) > 0 && !names[d.Name] {
			continue
		}
		src := expandPath(d.Src)
		dest := expandPath(d.Dest)
		out = append(out, entryStatus{
			name:  d.Name,
			kind:  "dotfile",
			state: dotfileState(src, dest),
			src:   src,
			dest:  dest,
		})
	}
	for _, r := range cfg.Repos {
		if len(names) > 0 && !names[r.Name] {
			continue
		}
		dest := expandPath(r.Dest)
		out = append(out, entryStatus{
			name:  r.Name,
			kind:  "repo",
			state: repoState(dest),
			url:   r.URL,
			dest:  dest,
		})
	}
	return out
}

func dotfileState(src, dest string) entryState {
	if _, err := os.Stat(src); err != nil {
		return stateSrcMissing
	}
	info, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return stateEmpty
	}
	if err != nil {
		return stateBlocked
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return stateBlocked
	}
	target, err := os.Readlink(dest)
	if err != nil {
		return stateBlocked
	}
	absSrc, _ := filepath.Abs(src)
	absTarget, _ := filepath.Abs(target)
	if absSrc != absTarget {
		return stateBlocked
	}
	return stateOK
}

func repoState(dest string) entryState {
	info, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return stateEmpty
	}
	if err != nil || !info.IsDir() {
		return stateBlocked
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		return stateOK
	}
	if _, err := os.Stat(filepath.Join(dest, ".jj")); err == nil {
		return stateOK
	}
	return stateBlocked
}

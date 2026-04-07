package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type entryState int

const (
	stateOK       entryState = iota // symlink exists and points to correct src
	stateMissing                    // nothing at dest
	stateConflict                   // real file exists at dest
	stateWrongTarget                // symlink exists but points elsewhere
	stateSrcMissing                 // src file doesn't exist in repo
	stateCloned                     // repo: dir exists and is a git repo
	stateNotCloned                  // repo: dest doesn't exist
)

func (s entryState) String() string {
	switch s {
	case stateOK:
		return "ok"
	case stateMissing:
		return "missing"
	case stateConflict:
		return "conflict"
	case stateWrongTarget:
		return "wrong-target"
	case stateSrcMissing:
		return "src-missing"
	case stateCloned:
		return "ok"
	case stateNotCloned:
		return "not-cloned"
	}
	return "unknown"
}

type entryStatus struct {
	name  string
	kind  string // "file" or "repo"
	state entryState
	src   string // dotfiles only
	dest  string
}

func cmdStatus(cfg *Config, args []string) int {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	nameOnly := fs.Bool("name-only", false, "print entry names only")
	conflicts := fs.Bool("conflicts", false, "print dest paths of conflicting entries only")
	fs.Parse(args)

	filter := map[string]bool{}
	for _, a := range fs.Args() {
		filter[a] = true
	}

	statuses := collectStatuses(cfg, filter)

	if *conflicts {
		for _, s := range statuses {
			if s.state == stateConflict {
				fmt.Println(s.dest)
			}
		}
		return 0
	}

	if *nameOnly {
		for _, s := range statuses {
			fmt.Println(s.name)
		}
		return 0
	}

	exit := 0
	for _, s := range statuses {
		fmt.Printf("%-20s %-6s %-12s %s\n", s.name, s.kind, s.state, s.dest)
		if s.state != stateOK && s.state != stateCloned {
			exit = 1
		}
	}
	return exit
}

func collectStatuses(cfg *Config, filter map[string]bool) []entryStatus {
	var out []entryStatus

	for _, d := range cfg.Dotfiles {
		if len(filter) > 0 && !filter[d.Name] {
			continue
		}
		src := expandPath(d.Src)
		dest := expandPath(d.Dest)
		out = append(out, entryStatus{
			name:  d.Name,
			kind:  "file",
			state: dotfileState(src, dest),
			src:   src,
			dest:  dest,
		})
	}

	for _, r := range cfg.Repos {
		if len(filter) > 0 && !filter[r.Name] {
			continue
		}
		dest := expandPath(r.Dest)
		out = append(out, entryStatus{
			name:  r.Name,
			kind:  "repo",
			state: repoState(dest),
			dest:  dest,
		})
	}

	return out
}

func dotfileState(src, dest string) entryState {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return stateSrcMissing
	}
	info, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return stateMissing
	}
	if err != nil {
		return stateConflict
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return stateConflict
	}
	target, err := os.Readlink(dest)
	if err != nil {
		return stateConflict
	}
	absSrc, _ := filepath.Abs(src)
	absTarget, _ := filepath.Abs(target)
	if absSrc != absTarget {
		return stateWrongTarget
	}
	return stateOK
}

func repoState(dest string) entryState {
	info, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return stateNotCloned
	}
	if err != nil || !info.IsDir() {
		return stateConflict
	}
	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		return stateCloned
	}
	// dir exists but no .git — could be jj or bare repo, treat as ok
	if _, err := os.Stat(filepath.Join(dest, ".jj")); err == nil {
		return stateCloned
	}
	return stateConflict
}
